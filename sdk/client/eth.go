package client

import (
	"math/big"
	"time"
)

type Balance *big.Int

var Wei Balance = big.NewInt(1)

var Gwei Balance = big.NewInt(1000000000)


func balance() Balance{
	 big.NewInt(100).Mul() time.Second
}