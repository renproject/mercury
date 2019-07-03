package main

import (
	"os"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/rpc/ethrpc"
	"github.com/renproject/mercury/rpc/zecrpc"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/renproject/mercury/types/zectypes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialise logger.
	logger := logrus.StandardLogger()

	// Retrieve data from environment variables.
	btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USER")
	btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
	zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
	zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USER")
	zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
	infuraAPIKey := os.Getenv("INFURA_KEY_DEFAULT")

	store := kv.NewJSON(kv.NewMemDB())
	cache := cache.New(store, logger)

	// Initialise Bitcoin API.
	btcTestnetNodeClient, err := btcrpc.NewNodeClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
	if err != nil {
		logger.Fatalf("cannot construct btc client: %v", err)
	}
	btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
	btcTestnetAPI := api.NewBtcApi(btctypes.Testnet, btcTestnetProxy, cache, logger)

	// Initialise ZCash API.
	zecTestnetNodeClient, err := zecrpc.NewNodeClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
	if err != nil {
		logger.Fatalf("cannot construct zec client: %v", err)
	}
	zecTestnetProxy := proxy.NewProxy(zecTestnetNodeClient)
	zecTestnetAPI := api.NewZecApi(zectypes.Testnet, zecTestnetProxy, cache, logger)

	// Initialize Ethereum API.
	infuraMainnetClient, err := ethrpc.NewInfuraClient(ethtypes.Mainnet, infuraAPIKey)
	if err != nil {
		logger.Fatalf("cannot construct infura mainnet client: %v", err)
	}
	ethMainnetProxy := proxy.NewProxy(infuraMainnetClient)
	ethMainnetAPI := api.NewEthApi(ethtypes.Mainnet, ethMainnetProxy, cache, logger)

	infuraTestnetClient, err := ethrpc.NewInfuraClient(ethtypes.Kovan, infuraAPIKey)
	if err != nil {
		logger.Fatalf("cannot construct infura testnet client: %v", err)
	}
	ethTestnetProxy := proxy.NewProxy(infuraTestnetClient)
	ethTestnetAPI := api.NewEthApi(ethtypes.Kovan, ethTestnetProxy, cache, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcTestnetAPI, zecTestnetAPI, ethMainnetAPI, ethTestnetAPI)
	server.Run()
}
