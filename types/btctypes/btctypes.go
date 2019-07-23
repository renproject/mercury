package btctypes

import (
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

type Network interface {
	types.Network

	Params() *chaincfg.Params
}

type network uint8

const (
	BtcLocalnet network = 0
	BtcMainnet  network = 1
	BtcTestnet  network = 2

	ZecLocalnet network = 3
	ZecMainnet  network = 4
	ZecTestnet  network = 5
)

// NewNetwork parse the network from a string.
func NewNetwork(chain types.Chain, network string) Network {
	switch chain {
	case types.Bitcoin:
		return NewBtcNetwork(network)
	case types.ZCash:
		return NewZecNetwork(network)
	default:
		panic(types.ErrUnknownChain)
	}
}

// NewBtcNetwork parse the btc network from a string.
func NewBtcNetwork(network string) Network {
	network = strings.ToLower(strings.TrimSpace(network))
	switch network {
	case "mainnet":
		return BtcMainnet
	case "testnet", "testnet3":
		return BtcTestnet
	case "localnet", "localhost":
		return BtcLocalnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// NewZecNetwork parse the zec network from a string.
func NewZecNetwork(network string) Network {
	network = strings.ToLower(strings.TrimSpace(network))
	switch network {
	case "mainnet":
		return ZecMainnet
	case "testnet", "testnet3":
		return ZecTestnet
	case "localnet", "localhost":
		return ZecLocalnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Params returns the params config for the network
func (network network) Params() *chaincfg.Params {
	switch network {
	case BtcMainnet, ZecMainnet:
		return &chaincfg.MainNetParams
	case BtcTestnet, BtcLocalnet, ZecTestnet, ZecLocalnet:
		return &chaincfg.TestNet3Params
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// String implements the `Stringer` interface.
func (network network) String() string {
	switch network {
	case BtcMainnet, ZecMainnet:
		return "mainnet"
	case BtcTestnet, ZecTestnet:
		return "testnet"
	case BtcLocalnet, ZecLocalnet:
		return "localnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Chain implements the types.Network interface.
func (network network) Chain() types.Chain {
	switch network {
	case BtcMainnet, BtcTestnet, BtcLocalnet:
		return types.Bitcoin
	case ZecMainnet, ZecTestnet, ZecLocalnet:
		return types.ZCash
	default:
		panic(types.ErrUnknownNetwork)
	}
}
