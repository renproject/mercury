package ethtypes

import (
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/renproject/mercury/types"
)

const (
	Mainnet Network = 1
	Kovan   Network = 42
)

func (network Network) String() string {
	switch network {
	case Mainnet:
		return "mainnet"
	case Kovan:
		return "kovan"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

type Network uint8

type USTX *coretypes.Transaction
type STX *coretypes.Transaction
type Address common.Address

func AddressFromHex(addr string) Address {
	return Address(common.HexToAddress(addr))
}

func (addr Address) Hex() string {
	return common.Address(addr).Hex()
}
