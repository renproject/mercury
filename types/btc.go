package types

import (
	"crypto/ecdsa"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// ErrUnknownBtcNetwork is returned when the given bitcoin network is unknwon to us.
var ErrUnknownBtcNetwork = errors.New("unknown bitcoin network")

//
type BtcValue int64

const (
	Satoshi BtcValue = 1
	BTC              = 100000000 * Satoshi
)

var (
	TestNet3Params = &chaincfg.TestNet3Params

	MainNetParams = &chaincfg.MainNetParams
)

//
type BtcNetwork uint8

const (
	BtcMainnet BtcNetwork = 1
	BtcTestnet BtcNetwork = 2
)

func (network BtcNetwork) Params() *chaincfg.Params {
	switch network {
	case BtcMainnet:
		return MainNetParams
	case BtcTestnet:
		return TestNet3Params
	default:
		panic(ErrUnknownBtcNetwork)
	}
}

func (network BtcNetwork) String() string {
	switch network {
	case BtcMainnet:
		return "mainnet"
	case BtcTestnet:
		return "testnet"
	default:
		panic(ErrUnknownBtcNetwork)
	}
}

//
type BtcAddr btcutil.Address

// AddressFromBase58String decode the base58 encoding bitcoin address to a `BtcAddr`.
func AddressFromBase58String(addr string, network BtcNetwork) (BtcAddr, error) {
	return btcutil.DecodeAddress(addr, network.Params())
}

func AddressFromPubKey(pubkey *ecdsa.PublicKey, network BtcNetwork) (BtcAddr, error) {
	return btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
}

func SerializePublicKey(pubKey *ecdsa.PublicKey, network BtcNetwork) []byte {
	switch network {
	case BtcMainnet:
		return (*btcec.PublicKey)(pubKey).SerializeCompressed()
	case BtcTestnet:
		return (*btcec.PublicKey)(pubKey).SerializeUncompressed()
	default:
		panic(ErrUnknownBtcNetwork)
	}
}

type TxHash string

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}
