package btcaddress

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/types/btctypes"
)

// BtcAddressFromBase58 decodes the `Address` from a Base58 encoding.
func BtcAddressFromBase58(addr string, network btctypes.Network) (Address, error) {
	return btcutil.DecodeAddress(addr, network.Params())
}

// BtcAddressFromPubKey decodes the `Address` from a public key.
func BtcAddressFromPubKey(pubkey *ecdsa.PublicKey, network btctypes.Network) (Address, error) {
	addr, err := btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
	if err != nil {
		return nil, fmt.Errorf("cannot decode address from public key: %v", err)
	}

	return btcutil.DecodeAddress(addr.EncodeAddress(), network.Params())
}
