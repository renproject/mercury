package btctypes

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
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
	SAT = Amount(1)
	BTC = Amount(1e8 * SAT)
)

// Network of Bitcoin blockchain.
type Network uint8

const (
	Localnet Network = 0
	Mainnet  Network = 1
	Testnet  Network = 2
)

// NewNetwork parse the network from a string.
func NewNetwork(network string) Network {
	network = strings.ToLower(strings.TrimSpace(network))
	switch network {
	case "mainnet":
		return Mainnet
	case "testnet", "testnet3":
		return Testnet
	case "localnet", "localhost":
		return Localnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Params returns the params config for the network
func (network Network) Params() *chaincfg.Params {
	switch network {
	case Mainnet:
		return &chaincfg.MainNetParams
	case Testnet, Localnet:
		return &chaincfg.TestNet3Params
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
	case Localnet:
		return "localnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Addr is an interface type for any type of destination a transaction output may spend to. This includes pay-to-pubkey
// (P2PK), pay-to-pubkey-hash (P2PKH), and pay-to-script-hash (P2SH). Address is designed to be generic enough that
// other kinds of addresses may be added in the future without changing the decoding and encoding API.
type Address btcutil.Address

// AddressFromBase58 decodes the base58 encoding bitcoin address to a `Addr`.
func AddressFromBase58(addr string, network Network) (Address, error) {
	return btcutil.DecodeAddress(addr, network.Params())
}

// AddressFromPubKey gets the `Addr` from a public key.
func AddressFromPubKey(pubkey *ecdsa.PublicKey, network Network) (Address, error) {
	addr, err := btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
	if err != nil {
		return nil, fmt.Errorf("cannot decode address from public key: %v", err)
	}

	return btcutil.DecodeAddress(addr.EncodeAddress(), network.Params())
}

// SerializePublicKey serializes the public key to bytes.
func SerializePublicKey(pubkey *ecdsa.PublicKey, network Network) []byte {
	switch network {
	case Mainnet:
		return (*btcec.PublicKey)(pubkey).SerializeCompressed()
	case Testnet, Localnet:
		return (*btcec.PublicKey)(pubkey).SerializeUncompressed()
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// RandomAddress returns a random Addr on given network.
func RandomAddress(network Network) (Address, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return AddressFromPubKey(&key.PublicKey, network)
}

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

type Confirmations int64
