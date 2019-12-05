package api

import "github.com/renproject/mercury/types"

func WhitelistLevel(network types.Network, method string) types.AccessLevel {
	switch network.Chain() {
	case types.Bitcoin, types.ZCash, types.BitcoinCash:
		return BtcWhitelistLevel(method)
	case types.Ethereum:
		return EthWhitelistLevel(method)
	default:
		return types.NoAccess
	}
}

func EthWhitelistLevel(method string) types.AccessLevel {
	switch method {
	case "eth_call":
		return types.CachedAccess
	default:
		return types.FullAccess
	}
}

func BtcWhitelistLevel(method string) types.AccessLevel {
	switch method {
	case "listunspent", "gettxout", "getrawtransaction":
		return types.FullAccess
	case "sendrawtransaction":
		return types.CachedAccess
	default:
		return types.NoAccess
	}
}
