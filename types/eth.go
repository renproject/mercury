package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type WEI big.Int

type ETH big.Float

type EthAddr = common.Address

var BytesToEthAddr = common.BytesToAddress
var HexStringToEthAddr = common.HexToAddress
