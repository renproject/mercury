package bnctypes

import "github.com/binance-chain/go-sdk/common/types"

type Coin types.Coin
type Coins types.Coins

func NewBNBCoin(amount int64) Coin {
	return Coin{"BNB", amount}
}

func NewRecipent(address Address, coins ...Coin) Recipient {
	return Recipient{address, NewCoins(coins...)}
}

func NewCoins(coins ...Coin) Coins {
	bCoins := make(Coins, len(coins))
	for i := range coins {
		bCoins[i] = types.Coin(coins[i])
	}
	return bCoins
}
