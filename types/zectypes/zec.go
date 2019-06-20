package zectypes

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

//
type ZecValue int64

const (
	Zatoshi ZecValue = 1
	ZEC              = 100000000 * Zatoshi
)

//
type Network uint8

const (
	Mainnet Network = 1
	Testnet Network = 2
)

func (network Network) Params() *chaincfg.Params {
	switch network {
	case Mainnet:
		return btctypes.MainNetParams
	case Testnet:
		return btctypes.TestNet3Params
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func (network Network) String() string {
	switch network {
	case Mainnet:
		return "mainnet"
	case Testnet:
		return "testnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

//
type Addr btcutil.Address

// AddressFromBase58String decodes the base58 encoding bitcoin address to a `Addr`.
func AddressFromBase58String(addr string, network Network) (Addr, error) {
	return btcutil.DecodeAddress(addr, network.Params())
}

// AddressFromPubKey gets the `Addr` from a public key.
func AddressFromPubKey(pubkey *ecdsa.PublicKey, network Network) (Addr, error) {
	return btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
}

// SerializePublicKey serializes the public key to bytes.
func SerializePublicKey(pubKey *ecdsa.PublicKey, network Network) []byte {
	switch network {
	case Mainnet:
		return (*btcec.PublicKey)(pubKey).SerializeCompressed()
	case Testnet:
		return (*btcec.PublicKey)(pubKey).SerializeUncompressed()
	default:
		panic(types.ErrUnknownNetwork)
	}
}

type TxHash string

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}
