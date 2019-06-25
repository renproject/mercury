package btctypes

import (
	"crypto/ecdsa"
	"crypto/rand"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/renproject/mercury/types"
)

// Amount represents bitcoin value in Satoshi (1e-8 Bitcoin).
type Amount int64

const (
	Satoshi Amount = 1
	Bitcoin        = 1e8 * Satoshi
)

// Network params for different networks.
var (
	TestNet3Params = &chaincfg.TestNet3Params

	MainNetParams = &chaincfg.MainNetParams
)

// Network of Ethereum blockchain.
type Network uint8

const (
	Mainnet Network = 1
	Testnet Network = 2
)

// NewNetwork parse the network from a string.
func NewNetwork(network string) Network {
	network = strings.ToLower(strings.TrimSpace(network))
	switch network {
	case "mainnet":
		return Mainnet
	case "testnet", "testnet3":
		return Testnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Params returns the params config for the network
func (network Network) Params() *chaincfg.Params {
	switch network {
	case Mainnet:
		return MainNetParams
	case Testnet:
		return TestNet3Params
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// String implements the `Stringer` interface.
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

// Addr is an interface type for any type of destination a transaction output may spend to. This includes pay-to-pubkey
// (P2PK), pay-to-pubkey-hash (P2PKH), and pay-to-script-hash (P2SH). Address is designed to be generic enough that
// other kinds of addresses may be added in the future without changing the decoding and encoding API.
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

// RandAddressPubKey returns a random Addr on given network.
func RandAddressPubKey(network Network) (Addr, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return AddressFromPubKey(&key.PublicKey,network)
}

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       Amount `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}
