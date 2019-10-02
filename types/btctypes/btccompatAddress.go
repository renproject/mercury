package btctypes

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes/bch"
	"golang.org/x/crypto/ripemd160"
)

type BtcCompatP2PKHAddress struct {
	pubKeyHash []byte
	network    Network
}

func NewAddressPubKey(pubKey []byte, network Network) btcutil.Address {
	return &BtcCompatP2PKHAddress{btcutil.Hash160(pubKey), network}
}

func NewAddressPubKeyHash(pubKeyHash []byte, network Network) btcutil.Address {
	return &BtcCompatP2PKHAddress{pubKeyHash, network}
}

func (address *BtcCompatP2PKHAddress) EncodeAddress() string {
	return EncodeAddress(P2PKH, address.pubKeyHash, address.network)
}

func (address *BtcCompatP2PKHAddress) String() string {
	return address.EncodeAddress()
}

func (address *BtcCompatP2PKHAddress) ScriptAddress() []byte {
	return address.pubKeyHash
}

func (address *BtcCompatP2PKHAddress) IsForNet(params *chaincfg.Params) bool {
	return address.network.Params().Name == params.Name
}

type BtcCompatP2SHAddress struct {
	scriptHash []byte
	network    Network
}

func NewAddressScriptHash(serializedScript []byte, network Network) btcutil.Address {
	return &BtcCompatP2SHAddress{btcutil.Hash160(serializedScript), network}
}

func NewAddressScriptHashFromHash(scriptHash []byte, network Network) btcutil.Address {
	return &BtcCompatP2PKHAddress{scriptHash, network}
}

func (address *BtcCompatP2SHAddress) EncodeAddress() string {
	return EncodeAddress(P2SH, address.scriptHash, address.network)
}

func (address *BtcCompatP2SHAddress) String() string {
	return address.EncodeAddress()
}

func (address *BtcCompatP2SHAddress) ScriptAddress() []byte {
	return address.scriptHash
}

func (address *BtcCompatP2SHAddress) IsForNet(params *chaincfg.Params) bool {
	return address.network.Params().Name == params.Name
}

func EncodeAddress(addrType AddressType, hash []byte, network Network) string {
	switch network := network.(type) {
	case ZecNetwork:
		return encodeAddress(hash, network.Prefix(addrType))
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func DecodeAddress(address string) (btcutil.Address, error) {
	var decoded = base58.Decode(address)
	if len(decoded) != 26 && len(decoded) != 25 {
		return nil, base58.ErrInvalidFormat
	}

	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])
	if checksum(decoded[:len(decoded)-4]) != cksum {
		return nil, base58.ErrChecksum
	}

	if len(decoded)-6 != ripemd160.Size && len(decoded)-5 != ripemd160.Size {
		return nil, errors.New("incorrect payload len")
	}

	var net Network
	var addrType AddressType
	var hash []byte
	if len(decoded) == 26 {
		addrType, net = ParsePrefix(decoded[:2])
		copy(hash, decoded[2:22])
	} else {
		addrType, net = ParsePrefix(decoded[:1])
		copy(hash, decoded[1:21])
	}

	switch addrType {
	case P2PKH:
		return &BtcCompatP2PKHAddress{
			pubKeyHash: hash,
			network:    net,
		}, nil
	case P2SH:
		return &BtcCompatP2SHAddress{
			scriptHash: hash,
			network:    net,
		}, nil
	}

	return nil, errors.New("unknown address")
}

// PayToAddrScript gets the PayToAddrScript for an address on the given blockchain
func PayToAddrScript(address Address, network Network) ([]byte, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return txscript.PayToAddrScript(address)
	case types.BitcoinCash:
		return bch.PayToAddrScript(address)
	}

	switch address := address.(type) {
	case *BtcCompatP2PKHAddress:
		return txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
			AddData(address.pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
			Script()
	case *BtcCompatP2SHAddress:
		return txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(address.scriptHash).
			AddOp(txscript.OP_EQUAL).Script()
	default:
		return nil, fmt.Errorf("unsupported address type %T", address)
	}
}

func encodeAddress(addrHash, prefix []byte) string {
	var (
		body  = append(prefix, addrHash...)
		chk   = checksum(body)
		cksum [4]byte
	)
	copy(cksum[:], chk[:4])
	return base58.Encode(append(body, cksum[:]...))
}

func checksum(input []byte) (cksum [4]byte) {
	var (
		h  = sha256.Sum256(input)
		h2 = sha256.Sum256(h[:])
	)
	copy(cksum[:], h2[:4])
	return
}
