package types

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Amount struct {
	value *big.Int
}

func (a Amount) Add(other Amount) Amount {
	v := big.NewInt(1)
	v.Add(a.value, other.value)
	return Amount{
		value: v,
	}
}

func (a Amount) Mul(other Amount) Amount {
	v := big.NewInt(1)
	v.Mul(a.value, other.value)
	return Amount{
		value: v,
	}
}

func NewAmount(bigWeiValue *big.Int) Amount {
	return Amount{
		value: bigWeiValue,
	}
}

func Wei(val uint64) Amount {
	return NewAmount(new(big.Int).SetUint64(val))
}

func Gwei(val uint64) Amount {
	return NewAmount(new(big.Int).SetUint64(val)).Mul(GWEI)
}

func Ether(val uint64) Amount {
	return NewAmount(new(big.Int).SetUint64(val)).Mul(ETHER)
}

var (
	WEI   = Wei(1)
	GWEI  = Wei(1000000000)
	ETHER = Wei(1000000000000000000)
)

const (
	EthMainnet EthNetwork = 1
	EthKovan   EthNetwork = 42
)

// ErrUnknownEthNetwork is returned when the given bitcoin network is unknwon to us.
var ErrUnknownEthNetwork = errors.New("unknown ethereum network")

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

type EthNetwork uint8

type EthAddr common.Address

func HexStringToEthAddr(addr string) EthAddr {
	return EthAddr(common.HexToAddress(addr))
}

func (addr EthAddr) Hex() string {
	return common.Address(addr).Hex()
}
