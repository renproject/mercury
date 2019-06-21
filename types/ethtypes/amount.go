package ethtypes

import "math/big"

type Amount struct {
	value *big.Int
}

func (a Amount) String() string {
	return a.value.String()
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

func (a Amount) Lte(other Amount) bool {
	return a.Lt(other) || a.Eq(other)
}

func (a Amount) Gte(other Amount) bool {
	return a.Gt(other) || a.Eq(other)
}

func (a Amount) Sub(other Amount) Amount {
	v := big.NewInt(1)
	v.Sub(a.value, other.value)
	return Amount{
		value: v,
	}
}

func (a Amount) Div(other Amount) Amount {
	v := big.NewInt(1)
	v.Div(a.value, other.value)
	return Amount{
		value: v,
	}
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

func newAmount(value *big.Int) Amount {
	return Amount{
		value: value,
	}
}

func WeiFromBig(weiAmount *big.Int) Amount {
	return newAmount(weiAmount)
}

func Wei(val uint64) Amount {
	return WeiFromBig(new(big.Int).SetUint64(val))
}

func GweiFromBig(gweiAmount *big.Int) Amount {
	return newAmount(gweiAmount).Mul(GWEI)
}

func Gwei(val uint64) Amount {
	return GweiFromBig(new(big.Int).SetUint64(val))
}

func EtherFromBig(etherAmount *big.Int) Amount {
	return newAmount(etherAmount).Mul(ETHER)
}

func Ether(val uint64) Amount {
	return EtherFromBig(new(big.Int).SetUint64(val))
}

var (
	WEI   = Wei(1)
	GWEI  = Wei(1000000000)
	ETHER = Wei(1000000000000000000)
)
