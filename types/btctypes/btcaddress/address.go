package btcaddress

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

// Address is an interface type for any type of destination a transaction output may spend to. This includes pay-to-
// pubkey (P2PK), pay-to-pubkey-hash (P2PKH), and pay-to-script-hash (P2SH). Address is designed to be generic enough
// that other kinds of addresses may be added in the future without changing the decoding and encoding API.
type Address btcutil.Address

type Recipient struct {
	Address Address
	Amount  btctypes.Amount
}

type Recipients []Recipient

// SerializePublicKey serializes the public key to bytes.
func SerializePublicKey(pubkey *ecdsa.PublicKey, network btctypes.Network) []byte {
	switch network {
	case btctypes.BtcMainnet, btctypes.ZecMainnet:
		return (*btcec.PublicKey)(pubkey).SerializeCompressed()
	case btctypes.BtcTestnet, btctypes.BtcLocalnet, btctypes.ZecTestnet, btctypes.ZecLocalnet:
		return (*btcec.PublicKey)(pubkey).SerializeUncompressed()
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// AddressFromBase58 decodes the base58 encoded address to an `Address`.
func AddressFromBase58(addr string, network btctypes.Network) (Address, error) {
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
func AddressFromPubKey(pubkey *ecdsa.PublicKey, network btctypes.Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		addr, err := btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
		if err != nil {
			return nil, fmt.Errorf("cannot decode address from public key: %v", err)
		}
		return btcutil.DecodeAddress(addr.EncodeAddress(), network.Params())
	case types.ZCash:
		return zecAddressFromHash160(btcutil.Hash160(SerializePublicKey(pubkey, network)), network.Params(), false)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// AddressFromScript gets the `Address` from a script.
func AddressFromScript(script []byte, network btctypes.Network) (Address, error) {
	switch network.Chain() {
	case types.Bitcoin:
		addr, err := btcutil.NewAddressScriptHash(script, network.Params())
		if err != nil {
			return nil, fmt.Errorf("cannot decode address from public key: %v", err)
		}
		return btcutil.DecodeAddress(addr.EncodeAddress(), network.Params())
	case types.ZCash:
		return zecAddressFromHash160(btcutil.Hash160(script), network.Params(), true)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}

// PayToAddrScript gets the PayToAddrScript for an address on the given blockchain
func PayToAddrScript(address Address, network btctypes.Network) ([]byte, error) {
	switch network.Chain() {
	case types.Bitcoin:
		return txscript.PayToAddrScript(address)
	case types.ZCash:
		return zecutil.PayToAddrScript(address)
	default:
		return nil, fmt.Errorf("unsupported blockchain: %v", network.Chain())
	}
}
