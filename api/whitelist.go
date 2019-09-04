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
	case "eth_gasPrice", "eth_blockNumber", "eth_getBalance", "eth_getBlockByNumber", "eth_getTransactionCount",
		"eth_call", "eth_estimateGas", "eth_pendingTransactions", "eth_getFilterChanges", "eth_getFilterLogs",
		"eth_getLogs", "eth_getWork", "eth_getProof":
		return types.FullAccess
	case "net_version", "eth_chainId", "eth_getBlockTransactionCountByHash", "eth_getBlockTransactionCountByNumber", "eth_getStorageAt",
		"eth_getUncleCountByBlockHash", "eth_getUncleCountByBlockNumber", "eth_getUncleByBlockHashAndIndex",
		"eth_getUncleByBlockNumberAndIndex", "eth_sign", "eth_getCode", "eth_sendTransaction", "eth_sendRawTransaction",
		"eth_getBlockByHash", "eth_getTransactionByHash", "eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex", "eth_getTransactionReceipt", "eth_newFilter", "eth_newBlockFilter",
		"eth_newPendingTransactionFilter", "eth_uninstallFilter", "eth_submitWork", "eth_submitHashrate":
		return types.CachedAccess
	default:
		return types.NoAccess
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
