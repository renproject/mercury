package main

import (
	"os"
	"time"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/rpc/ethrpc"
	"github.com/renproject/mercury/rpc/zecrpc"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialise logger.
	logger := logrus.StandardLogger()

	// Initialise stores.
	ethKovanStore := kv.NewTTLCache(kv.NewJSON(kv.NewMemDB()), 10*time.Second)
	ethKovanCache := cache.New(ethKovanStore, logger)
	ethStore := kv.NewTTLCache(kv.NewJSON(kv.NewMemDB()), 10*time.Second)
	ethCache := cache.New(ethStore, logger)

	btcStore := kv.NewTTLCache(kv.NewJSON(kv.NewMemDB()), 10*time.Second)
	btcCache := cache.New(btcStore, logger)

	zecStore := kv.NewTTLCache(kv.NewJSON(kv.NewMemDB()), 10*time.Second)
	zecCache := cache.New(zecStore, logger)

	// Initialise Bitcoin API.
	btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
	btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
	btcTestnetNodeClient, err := btcrpc.NewNodeClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
	if err != nil {
		logger.Fatalf("cannot construct btc client: %v", err)
	}
	btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
	btcTestnetAPI := api.NewBtcApi(btctypes.BtcTestnet, btcTestnetProxy, btcCache, logger)

	// Initialise ZCash API.
	zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
	zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
	zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
	zecTestnetNodeClient, err := zecrpc.NewNodeClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
	if err != nil {
		logger.Fatalf("cannot construct zec client: %v", err)
	}
	zecTestnetProxy := proxy.NewProxy(zecTestnetNodeClient)
	zecTestnetAPI := api.NewZecApi(btctypes.ZecTestnet, zecTestnetProxy, zecCache, logger)

	// Initialize Ethereum API.
	infuraAPIKey := os.Getenv("INFURA_KEY_DEFAULT")
	taggedKeys := map[string]string{
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	infuraMainnetClient, err := ethrpc.NewInfuraClient(ethtypes.Mainnet, infuraAPIKey, taggedKeys)
	if err != nil {
		logger.Fatalf("cannot construct infura mainnet client: %v", err)
	}
	ethMainnetProxy := proxy.NewProxy(infuraMainnetClient)
	ethMainnetAPI := api.NewEthApi(ethtypes.Mainnet, ethMainnetProxy, ethCache, logger)

	infuraTestnetClient, err := ethrpc.NewInfuraClient(ethtypes.Kovan, infuraAPIKey, taggedKeys)
	if err != nil {
		logger.Fatalf("cannot construct infura testnet client: %v", err)
	}
	ethTestnetProxy := proxy.NewProxy(infuraTestnetClient)
	ethTestnetAPI := api.NewEthApi(ethtypes.Kovan, ethTestnetProxy, ethKovanCache, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcTestnetAPI, zecTestnetAPI, ethMainnetAPI, ethTestnetAPI)
	server.Run()
}
