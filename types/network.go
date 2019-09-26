package types

import (
	"fmt"
	"strings"
)

type Chain uint8

const (
	Bitcoin     Chain = 0
	Ethereum    Chain = 1
	ZCash       Chain = 2
	BitcoinCash Chain = 3
	Binance     Chain = 4
)

func NewChain(chain string) Chain {
	chain = strings.ToUpper(chain)
	switch chain {
	case "BITCOIN", "BTC":
		return Bitcoin
	case "ETHEREUM", "ETH":
		return Ethereum
	case "ZCASH", "ZEC":
		return ZCash
	case "BITCOIN CASH", "BCH", "BITCOINCASH":
		return BitcoinCash
	case "BINANCE", "BNC", "BNB":
		return Binance
	default:
		panic(ErrUnknownChain)
	}
}

// String implements the `Stringer` interface.
func (chain Chain) String() string {
	switch chain {
	case Bitcoin:
		return "btc"
	case Ethereum:
		return "eth"
	case ZCash:
		return "zec"
	case BitcoinCash:
		return "bch"
	case Binance:
		return "bnc"
	default:
		panic(ErrUnknownChain)
	}
}

// Network of the blockchain.
type Network interface {
	fmt.Stringer
	Chain() Chain
}
