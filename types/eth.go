package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

type EthAddr = common.Address

var (
	Wei   = big.NewInt(1)
	Kwei  = mul(math.BigPow(10, 3), Wei)
	Mwei  = mul(math.BigPow(10, 6), Wei)
	Gwei  = mul(math.BigPow(10, 9), Wei)
	Ether = mul(math.BigPow(10, 18), Wei)
)

var BytesToEthAddr = common.BytesToAddress
var HexStringToEthAddr = common.HexToAddress

func mul(x, y *big.Int) *big.Int {
	result := big.NewInt(1)
	result.Mul(x, y)
	return result
}
