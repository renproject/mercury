package ethtypes

import (
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/renproject/mercury/types"
)

const (
	EthMainnet EthNetwork = 1
	EthKovan   EthNetwork = 42
)

func (network EthNetwork) String() string {
	switch network {
	case EthMainnet:
		return "mainnet"
	case EthKovan:
		return "kovan"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

type EthNetwork uint8

type EthSignedTx *coretypes.Transaction
type EthAddr common.Address

func HexStringToEthAddr(addr string) EthAddr {
	return EthAddr(common.HexToAddress(addr))
}

func (addr EthAddr) Hex() string {
	return common.Address(addr).Hex()
}
