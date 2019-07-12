package btcaddress

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/iqoption/zecutil"
	"github.com/renproject/mercury/types/btctypes"
)

// ZecAddressFromBase58 decodes the base58 encoding ZCash address to an `Address`.
func ZecAddressFromBase58(addr string, network btctypes.Network) (Address, error) {
	return zecutil.DecodeAddress(addr, network.Params().Name)
}

// ZecAddressFromPubKey gets the `Address` from a public key.
func ZecAddressFromPubKey(pubkey *ecdsa.PublicKey, network btctypes.Network) (Address, error) {
	addr, err := addressFromHash160(btcutil.Hash160(SerializePublicKey(pubkey, network)), network.Params(), false)
	if err != nil {
		return nil, fmt.Errorf("cannot decode address from public key: %v", err)
	}

	return zecutil.DecodeAddress(addr.EncodeAddress(), network.Params().Name)
}

func addressFromHash160(hash []byte, params *chaincfg.Params, isScript bool) (btcutil.Address, error) {
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
