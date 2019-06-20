package eth

import "math/big"

type Amount struct {
	value *big.Int
}

func (a Amount) Gt(other Amount) bool {
	return a.value.Cmp(other.value) > 0
}

func (a Amount) Lt(other Amount) bool {
	return a.value.Cmp(other.value) < 0
}

func (a Amount) Eq(other Amount) bool {
	return a.value.Cmp(other.value) == 0
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
