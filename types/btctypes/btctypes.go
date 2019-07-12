package btctypes

import (
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

// Amount represents the value in the smallest possible unit for the respective blockchain.
type Amount int64

const (
	SAT = Amount(1)
	BTC = Amount(1e8 * SAT)
)

const (
	ZAT = Amount(1)
	ZEC = Amount(1e8 * ZAT)
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
