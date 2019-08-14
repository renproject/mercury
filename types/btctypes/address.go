package btctypes

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/iqoption/zecutil"
	"github.com/renproject/mercury/types"
)

// Address is an interface type for any type of destination a transaction output may spend to. This includes pay-to-
// pubkey (P2PK), pay-to-pubkey-hash (P2PKH), and pay-to-script-hash (P2SH). Address is designed to be generic enough
// that other kinds of addresses may be added in the future without changing the decoding and encoding API.
type Address btcutil.Address

type Recipient struct {
	Address Address
	Amount  Amount
}

// NewRecipient creates a new recipient
func NewRecipient(address Address, amount Amount) Recipient {
	return Recipient{address, amount}
}

type Recipients []Recipient

// SerializePublicKey serializes the public key to bytes.
func SerializePublicKey(pubkey ecdsa.PublicKey, network Network) []byte {
	switch network {
	case BtcMainnet, ZecMainnet:
		return (*btcec.PublicKey)(&pubkey).SerializeCompressed()
	case BtcTestnet, BtcLocalnet, ZecTestnet, ZecLocalnet:
		return (*btcec.PublicKey)(&pubkey).SerializeUncompressed()
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// AddressFromBase58 decodes the base58 encoded address to an `Address`.
func AddressFromBase58(addr string, network Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return btcutil.DecodeAddress(addr, network.Params())
	case types.ZCash:
		return zecutil.DecodeAddress(addr, network.Params().Name)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// AddressFromPubKey gets the `Address` from a public key.
func AddressFromPubKey(pubkey ecdsa.PublicKey, network Network, segwit bool) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		if segwit {
			return btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(SerializePublicKey(pubkey, network)), network.Params())
		}
		return btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
	case types.ZCash:
		return zecAddressFromHash160(btcutil.Hash160(SerializePublicKey(pubkey, network)), network.Params(), false)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// AddressFromScript gets the `Address` from a script.
func AddressFromScript(script []byte, network Network, segwit bool) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		if segwit {
			scriptHash := sha256.Sum256(script)
			return btcutil.NewAddressWitnessScriptHash(scriptHash[:], network.Params())
		}
		return btcutil.NewAddressScriptHash(script, network.Params())
	case types.ZCash:
		return zecAddressFromHash160(btcutil.Hash160(script), network.Params(), true)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// PayToAddrScript gets the PayToAddrScript for an address on the given blockchain
func PayToAddrScript(address Address, network Network) ([]byte, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return txscript.PayToAddrScript(address)
	case types.ZCash:
		return zecutil.PayToAddrScript(address)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

func zecAddressFromHash160(hash []byte, params *chaincfg.Params, isScript bool) (btcutil.Address, error) {
	prefixes := map[string]map[string][]byte{
		"mainnet": map[string][]byte{
			"pubkey": []byte{0x1C, 0xB8},
			"script": []byte{0x1C, 0xBD},
		},
		"testnet3": map[string][]byte{
			"pubkey": []byte{0x1D, 0x25},
			"script": []byte{0x1C, 0xBA},
		},
		"regtest": map[string][]byte{
			"pubkey": []byte{0x1D, 0x25},
			"script": []byte{0x1C, 0xBA},
		},
	}
	if isScript {
		return zecutil.DecodeAddress(encodeHash(hash[:], prefixes[params.Name]["script"]), params.Name)
	}
	return zecutil.DecodeAddress(encodeHash(hash[:], prefixes[params.Name]["pubkey"]), params.Name)
}

func encodeHash(addrHash, prefix []byte) string {
	var (
		body  = append(prefix, addrHash...)
		chk   = addrChecksum(body)
		cksum [4]byte
	)
	copy(cksum[:], chk[:4])
	return base58.Encode(append(body, cksum[:]...))
}

func addrChecksum(input []byte) (cksum [4]byte) {
	var (
		h  = sha256.Sum256(input)
		h2 = sha256.Sum256(h[:])
	)
	copy(cksum[:], h2[:4])
	return
}
