package btctypes

import (
	"github.com/btcsuite/btcutil"
)

// Amount represents the value in the smallest possible unit for the respective blockchain.
type Amount int64

const (
	SAT = Amount(1)
	BTC = Amount(1e8 * SAT)
)

const (
	ZAT = Amount(1)
	ZEC = Amount(1e8 * ZAT)
)

func AmountFromFloat64(amount float64) Amount {
	amt, err := btcutil.NewAmount(amount)
	if err != nil {
		panic(err)
	}
	return Amount(amt.ToUnit(btcutil.AmountSatoshi))
}
