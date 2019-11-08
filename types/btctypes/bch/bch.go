package bch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/bech32"
	"golang.org/x/crypto/ripemd160"
)

type Version byte

const (
	P2PKH = Version(0x0)
	P2SH  = Version(0x8)
)

// BCashCodec is the BitcoinCash address encoding format according to
// `https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md#payload`
var BCashCodec = NewCodec("qpzry9x8gf2tvdw0s3jn54khce6mua7l")

type P2PKHAddress struct {
	params     *chaincfg.Params
	pubKeyHash []byte
}

func NewAddressPubKey(pubKey []byte, params *chaincfg.Params) btcutil.Address {
	return &P2PKHAddress{params, btcutil.Hash160(pubKey)}
}

func NewAddressPubKeyHash(pubKeyHash []byte, params *chaincfg.Params) btcutil.Address {
	return &P2PKHAddress{params, pubKeyHash}
}

func (address *P2PKHAddress) EncodeAddress() string {
	addr, err := EncodeAddress(P2PKH, address.pubKeyHash, address.params)
	if err != nil {
		panic(fmt.Sprintf("invariant violation: failed to encode address: %v", err))
	}
	return addr
}

func (address *P2PKHAddress) String() string {
	return address.EncodeAddress()
}

func (address *P2PKHAddress) ScriptAddress() []byte {
	return address.pubKeyHash
}

func (address *P2PKHAddress) IsForNet(params *chaincfg.Params) bool {
	return address.params.Name == params.Name
}

type P2SHAddress struct {
	params     *chaincfg.Params
	scriptHash []byte
}

func NewAddressScriptHash(serializedScript []byte, params *chaincfg.Params) btcutil.Address {
	return &P2SHAddress{params, btcutil.Hash160(serializedScript)}
}

func NewAddressScriptHashFromHash(scriptHash []byte, params *chaincfg.Params) btcutil.Address {
	return &P2PKHAddress{params, scriptHash}
}

func (address *P2SHAddress) EncodeAddress() string {
	addr, err := EncodeAddress(P2SH, address.scriptHash, address.params)
	if err != nil {
		panic(fmt.Sprintf("invariant violation: failed to encode address: %v", err))
	}
	return addr
}

func (address *P2SHAddress) String() string {
	return address.EncodeAddress()
}

func (address *P2SHAddress) ScriptAddress() []byte {
	return address.scriptHash
}

func (address *P2SHAddress) IsForNet(params *chaincfg.Params) bool {
	return address.params.Name == params.Name
}

// PolyMod implements `PolyMod` function
// uint64_t PolyMod(const data &v) {
// 	uint64_t c = 1;
// 	for (uint8_t d : v) {
// 		uint8_t c0 = c >> 35;
// 		c = ((c & 0x07ffffffff) << 5) ^ d;
// 		if (c0 & 0x01) c ^= 0x98f2bc8e61;
// 		if (c0 & 0x02) c ^= 0x79b76d99e2;
// 		if (c0 & 0x04) c ^= 0xf33e5fb3c4;
// 		if (c0 & 0x08) c ^= 0xae2eabe2a8;
// 		if (c0 & 0x10) c ^= 0x1e4f43e470;
// 	}
// 	return c ^ 1;
// }
// defined in `https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md`,
// it is  used to calculate the checksum for bitcoin cash addresses.
func PolyMod(v []byte) uint64 {
	c := uint64(1)
	for _, d := range v {
		c0 := byte(c >> 35)
		c = ((c & 0x07ffffffff) << 5) ^ uint64(d)

		if c0&0x01 > 0 {
			c ^= 0x98f2bc8e61
		}
		if c0&0x02 > 0 {
			c ^= 0x79b76d99e2
		}
		if c0&0x04 > 0 {
			c ^= 0xf33e5fb3c4
		}
		if c0&0x08 > 0 {
			c ^= 0xae2eabe2a8
		}
		if c0&0x10 > 0 {
			c ^= 0x1e4f43e470
		}
	}
	return c ^ 1
}

// EncodePrefix encodes prefix according to
// `https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md#checksum`
func EncodePrefix(prefixString string) []byte {
	prefixBytes := make([]byte, len(prefixString)+1)
	for i := 0; i < len(prefixString); i++ {
		prefixBytes[i] = byte(prefixString[i]) & 0x1f
	}
	prefixBytes[len(prefixString)] = 0
	return prefixBytes
}

// VerifyChecksum verifies whether the given payload is wellformed, according to
// `https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md#checksum`
func VerifyChecksum(prefix string, payload []byte) bool {
	return PolyMod(append(EncodePrefix(prefix), payload...)) == 0
}

// AppendChecksum creates a checksum for the given payload and appends it, according to
// `https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md#checksum`
func AppendChecksum(prefix string, payload []byte) []byte {
	prefixedPayload := append(EncodePrefix(prefix), payload...)

	// Append 8 zeroes.
	prefixedPayload = append(prefixedPayload, 0, 0, 0, 0, 0, 0, 0, 0)

	// Determine what to XOR into those 8 zeroes.
	mod := PolyMod(prefixedPayload)

	checksum := make([]byte, 8)
	for i := 0; i < 8; i++ {
		// Convert the 5-bit groups in mod to checksum values.
		checksum[i] = byte((mod >> uint(5*(7-i))) & 0x1f)
	}
	return append(payload, checksum...)
}

func DecodeAddress(addr string, params *chaincfg.Params) (btcutil.Address, error) {
	// Legacy address decoding
	if address, err := btcutil.DecodeAddress(addr, params); err == nil {
		switch address.(type) {
		case *btcutil.AddressPubKeyHash, *btcutil.AddressScriptHash, *btcutil.AddressPubKey:
			return address, nil
		case *btcutil.AddressWitnessPubKeyHash, *btcutil.AddressWitnessScriptHash:
			return nil, fmt.Errorf("bitcoin cash does not support SegWit addresses")
		default:
			return nil, fmt.Errorf("unsuported legacy bitcoin address type: %T", address)
		}
	}

	if addrParts := strings.Split(addr, ":"); len(addrParts) != 1 {
		addr = addrParts[1]
	}

	decoded := BCashCodec.DecodeString(addr)
	if !VerifyChecksum(prefix(params), decoded) {
		return nil, btcutil.ErrChecksumMismatch
	}

	addrBytes, err := bech32.ConvertBits(decoded[:len(decoded)-8], 5, 8, false)
	if err != nil {
		return nil, err
	}

	switch len(addrBytes) - 1 {
	case ripemd160.Size: // P2PKH or P2SH
		switch Version(addrBytes[0]) {
		case P2PKH:
			return &P2PKHAddress{params, addrBytes[1:21]}, nil
		case P2SH:
			return &P2SHAddress{params, addrBytes[1:21]}, nil
		default:
			return nil, btcutil.ErrUnknownAddressType
		}
	default:
		return nil, errors.New("decoded address is of unknown size")
	}
}

// PayToAddrScript creates a new script to pay a transaction output to a the
// specified address.
func PayToAddrScript(addr btcutil.Address) ([]byte, error) {
	if addr == nil {
		return nil, fmt.Errorf("unable to generate payment script for nil address")
	}
	switch addr := addr.(type) {
	case *P2PKHAddress:
		return txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
			AddData(addr.pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
			Script()
	case *P2SHAddress:
		return txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(addr.scriptHash).
			AddOp(txscript.OP_EQUAL).Script()
	case *btcutil.AddressPubKeyHash, *btcutil.AddressScriptHash, *btcutil.AddressPubKey:
		return txscript.PayToAddrScript(addr)
	default:
		return nil, fmt.Errorf("unsupported address type %T", addr)
	}
}

func EncodeAddress(ver Version, hash []byte, params *chaincfg.Params) (string, error) {
	if (len(hash)-20)/4 != int(ver)%8 {
		return "", fmt.Errorf("invalid version: %d", ver)
	}
	addrBytes, err := bech32.ConvertBits(append([]byte{byte(ver)}, hash...), 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to encode using bech32: %v", err)
	}
	return BCashCodec.EncodeToString(AppendChecksum(prefix(params), addrBytes)), nil
}

func prefix(params *chaincfg.Params) string {
	switch params.Name {
	case "mainnet":
		return "bitcoincash"
	case "testnet3":
		return "bchtest"
	case "regtest":
		return "bchreg"
	default:
		panic("unknown network")
	}
}
