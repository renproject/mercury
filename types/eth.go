package types

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// ErrUnknownEthNetwork is returned when the given bitcoin network is unknwon to us.
var ErrUnknownEthNetwork = errors.New("unknown ethereum network")

type EthNetwork uint8
type WeiValue = *big.Int

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
		panic(ErrUnknownEthNetwork)
	}
}

type EthAddr = common.Address

var (
	Wei   WeiValue = big.NewInt(1)
	Kwei  WeiValue = mul(math.BigPow(10, 3), Wei)
	Mwei  WeiValue = mul(math.BigPow(10, 6), Wei)
	Gwei  WeiValue = mul(math.BigPow(10, 9), Wei)
	Ether WeiValue = mul(math.BigPow(10, 18), Wei)
)

var BytesToEthAddr = common.BytesToAddress
var HexStringToEthAddr = common.HexToAddress

func mul(x, y *big.Int) *big.Int {
	result := big.NewInt(1)
	result.Mul(x, y)
	return result
}
