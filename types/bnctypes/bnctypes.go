package bnctypes

import (
	"strings"

	btypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/renproject/mercury/types"
)

type Network interface {
	types.Network

	ChainID() string
	ChainNetwork() btypes.ChainNetwork
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

func (net network) ChainID() string {
	switch net {
	case Testnet:
		return "Binance-Chain-Nile"
	case Mainnet:
		return "Binance-Chain-Tigris"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func (net network) Chain() types.Chain {
	return types.Binance
}

func (net network) ChainNetwork() btypes.ChainNetwork {
	switch net {
	case Testnet:
		return btypes.TestNetwork
	case Mainnet:
		return btypes.ProdNetwork
	default:
		panic(types.ErrUnknownNetwork)
	}
}
