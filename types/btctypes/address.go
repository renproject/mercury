package btctypes

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes/bch"
)

// Address is an interface type for any type of destination a transaction output may spend to. This includes pay-to-
// pubkey (P2PK), pay-to-pubkey-hash (P2PKH), and pay-to-script-hash (P2SH). Address is designed to be generic enough
// that other kinds of addresses may be added in the future without changing the decoding and encoding API.
type Address btcutil.Address

type AddressType uint32

const (
	P2PKH = AddressType(0)
	P2SH  = AddressType(1)
)

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
func SerializePublicKey(pubkey ecdsa.PublicKey) []byte {
	return (*btcec.PublicKey)(&pubkey).SerializeCompressed()
}

// AddressFromBase58 decodes the base58 encoded address to an `Address`.
func AddressFromBase58(addr string, network Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return btcutil.DecodeAddress(addr, network.Params())
	case types.ZCash:
		return DecodeAddress(addr)
	case types.BitcoinCash:
		return bch.DecodeAddress(addr, network.Params())
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// AddressFromPubKey gets the `Address` from a public key.
func AddressFromPubKey(pubkey ecdsa.PublicKey, network Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return btcutil.NewAddressPubKeyHash(btcutil.Hash160(SerializePublicKey(pubkey)), network.Params())
	case types.ZCash:
		return NewAddressPubKey(SerializePublicKey(pubkey), network), nil
	case types.BitcoinCash:
		return bch.NewAddressPubKey(SerializePublicKey(pubkey), network.Params()), nil
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// AddressFromPubKeyHash gets the `Address` from a public key hash.
func AddressFromPubKeyHash(pHash []byte, network Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return btcutil.NewAddressPubKeyHash(pHash, network.Params())
	case types.ZCash:
		return NewAddressPubKeyHash(pHash, network), nil
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// AddressFromScript gets the `Address` from a script.
func AddressFromScript(script []byte, network Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return btcutil.NewAddressScriptHash(script, network.Params())
	case types.ZCash:
		return NewAddressScriptHash(script, network), nil
	case types.BitcoinCash:
		return bch.NewAddressScriptHash(script, network.Params()), nil
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// SegWitAddressFromPubKey gets the SegWit compatible `Address` from a PubKey
func SegWitAddressFromPubKey(pubKey ecdsa.PublicKey, network Network) (Address, error) {
	if !network.SegWitEnabled() {
		return nil, ErrDoesNotSupportSegWit
	}
	switch network.Chain() {
	case types.Bitcoin:
		return btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(SerializePublicKey(pubKey)), network.Params())
	default:
		return nil, types.ErrUnknownChain
	}
}

// SegWitAddressFromScript gets the SegWit compatible `Address` from a Script.
func SegWitAddressFromScript(script []byte, network Network) (Address, error) {
	if !network.SegWitEnabled() {
		return nil, ErrDoesNotSupportSegWit
	}
	switch network.Chain() {
	case types.Bitcoin:
		scriptHash := sha256.Sum256(script)
		return btcutil.NewAddressWitnessScriptHash(scriptHash[:], network.Params())
	default:
		return nil, types.ErrUnknownChain
	}
}

var ErrUnknownAddressType = fmt.Errorf("unknown address type")
