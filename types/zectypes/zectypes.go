package zectypes

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

type Amount = btctypes.Amount

const (
	ZAT = Amount(1)
	ZEC = Amount(1e8 * ZAT)
)

type Network = btctypes.Network

const (
	Localnet = btctypes.Localnet
	Mainnet  = btctypes.Mainnet
	Testnet  = btctypes.Testnet
)

var NewNetwork = btctypes.NewNetwork

type Address = btctypes.Address

// AddressFromBase58 decodes the base58 encoding ZCash address to an `Address`.
func AddressFromBase58(addr string, network Network) (Address, error) {
	return zecutil.DecodeAddress(addr, network.Params().Name)
}

// AddressFromPubKey gets the `Address` from a public key.
func AddressFromPubKey(pubkey *ecdsa.PublicKey, network Network) (Address, error) {
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

var SerializePublicKey = btctypes.SerializePublicKey

type UTXO struct {
	TxHash       TxHash `json:"txHash"`
	Amount       Amount `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}

type UTXOs []UTXO

func (utxos *UTXOs) Sum() Amount {
	total := Amount(0)
	for _, utxo := range *utxos {
		total += Amount(utxo.Amount)
	}
	return total
}

type Recipient struct {
	Address Address
	Amount  Amount
}

type Recipients []Recipient

type Confirmations = btctypes.Confirmations
