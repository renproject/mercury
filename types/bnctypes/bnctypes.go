package bnctypes

import (
	"strings"

	"github.com/renproject/mercury/types"
)

type Network interface {
	types.Network

	ChainID() string
}

type network uint8

const (
	Testnet = network(0)
	Mainnet = network(1)
)

// NewNetwork parse the network from a string.
func NewNetwork(network string) Network {
	network = strings.ToLower(network)
	switch network {
	case "testnet":
		return Testnet
	case "mainnet":
		return Mainnet
	default:
		panic(types.ErrUnknownChain)
	}
}

func (net network) String() string {
	switch net {
	case Testnet:
		return "testnet"
	case Mainnet:
		return "mainnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func (net network) Chain() types.Chain {
	return types.Binance
}

func (net network) ChainID() string {
	switch net {
	case Testnet:
		return "testnet"
	case Mainnet:
		return "mainnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}
