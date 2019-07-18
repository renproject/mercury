package btcaddress

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/iqoption/zecutil"
)

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
