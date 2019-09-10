package btctypes

import (
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

type Network interface {
	types.Network

	Prefix(addrType AddressType) []byte
	Params() *chaincfg.Params
	SegWitEnabled() bool
}

type network uint8

const (
	BtcLocalnet network = 0
	BtcMainnet  network = 1
	BtcTestnet  network = 2

	BchLocalnet network = 6
	BchMainnet  network = 7
	BchTestnet  network = 8
)

// NewNetwork parse the network from a string.
func NewNetwork(chain types.Chain, network string) Network {
	switch chain {
	case types.Bitcoin:
		return NewBtcNetwork(network)
	case types.ZCash:
		return NewZecNetwork(network)
	case types.BitcoinCash:
		return NewBchNetwork(network)
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

// NewBchNetwork parse the ltc network from a string.
func NewBchNetwork(network string) Network {
	network = strings.ToLower(strings.TrimSpace(network))
	switch network {
	case "mainnet":
		return BchMainnet
	case "testnet", "testnet3":
		return BchTestnet
	case "localnet", "localhost":
		return BchLocalnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Params returns the params config for the network
func (network network) Params() *chaincfg.Params {
	switch network {
	case BtcMainnet, BchMainnet:
		return &chaincfg.MainNetParams
	case BtcTestnet, BtcLocalnet, BchTestnet, BchLocalnet:
		return &chaincfg.TestNet3Params
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// SegWitEnabled returns the params config for the network
func (network network) SegWitEnabled() bool {
	switch network.Chain() {
	case types.Bitcoin:
		return true
	case types.ZCash, types.BitcoinCash:
		return false
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// String implements the `Stringer` interface.
func (network network) String() string {
	switch network {
	case BtcMainnet, BchMainnet:
		return "mainnet"
	case BtcTestnet, BchTestnet:
		return "testnet"
	case BtcLocalnet, BchLocalnet:
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
	case BchMainnet, BchTestnet, BchLocalnet:
		return types.BitcoinCash
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Prefix implements the types.Network interface.
func (network network) Prefix(addrType AddressType) []byte {
	panic("unimplemented")
}
