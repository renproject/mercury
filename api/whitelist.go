package api

import "github.com/renproject/mercury/types"

func WhitelistLevel(network types.Network, method string) int64 {
	switch network.Chain() {
	case types.Bitcoin, types.ZCash:
		return BtcWhitelistLevel(method)
	case types.Ethereum:
		return EthWhitelistLevel(method)
	default:
		return 0
	}
}

func EthWhitelistLevel(method string) int64 {
	switch method {
	case "eth_gasPrice", "eth_blockNumber", "eth_getBalance", "eth_getTransactionCount", "eth_call", "eth_estimateGas",
		"eth_pendingTransactions", "eth_getFilterChanges", "eth_getFilterLogs", "eth_getLogs", "eth_getWork", "eth_getProof":
		return 2
	case "eth_getBlockTransactionCountByHash", "eth_getBlockTransactionCountByNumber", "eth_getStorageAt",
		"eth_getUncleCountByBlockHash", "eth_getUncleCountByBlockNumber", "eth_getUncleByBlockHashAndIndex",
		"eth_getUncleByBlockNumberAndIndex", "eth_sign", "eth_getCode", "eth_sendTransaction", "eth_sendRawTransaction",
		"eth_getBlockByHash", "eth_getBlockByNumber", "eth_getTransactionByHash", "eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex", "eth_getTransactionReceipt", "eth_newFilter", "eth_newBlockFilter",
		"eth_newPendingTransactionFilter", "eth_uninstallFilter", "eth_submitWork", "eth_submitHashrate":
		return 1
	default:
		return 0
	}
}

func BtcWhitelistLevel(method string) int64 {
	switch method {
	case "listunspent":
		return 2
	case "gettxout", "sendrawtransaction", "getrawtransaction":
		return 1
	default:
		return 0
	}
}
