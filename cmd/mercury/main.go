package main

import (
	"os"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialise logger.
	logger := logrus.StandardLogger()

	// Initialise stores.
	ethKovanStore := kv.NewJSON(kv.NewMemDB())
	ethKovanCache := cache.New(ethKovanStore, logger)
	ethStore := kv.NewJSON(kv.NewMemDB())
	ethCache := cache.New(ethStore, logger)
	btcStore := kv.NewJSON(kv.NewMemDB())
	btcCache := cache.New(btcStore, logger)
	zecStore := kv.NewJSON(kv.NewMemDB())
	zecCache := cache.New(zecStore, logger)
	bchStore := kv.NewJSON(kv.NewMemDB())
	bchCache := cache.New(bchStore, logger)
	maticStore := kv.NewJSON(kv.NewMemDB())
	maticCache := cache.New(maticStore, logger)

	// Initialise Bitcoin API.
	btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
	btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
	btcTestnetNodeClient := rpc.NewClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
	btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
	btcTestnetAPI := api.NewApi(btctypes.BtcTestnet, btcTestnetProxy, btcCache, logger)

	// Initialise ZCash API.
	zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
	zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
	zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
	zecTestnetNodeClient := rpc.NewClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
	zecTestnetProxy := proxy.NewProxy(zecTestnetNodeClient)
	zecTestnetAPI := api.NewApi(btctypes.ZecTestnet, zecTestnetProxy, zecCache, logger)

	// Initialise BCash API.
	bchTestnetURL := os.Getenv("BCASH_TESTNET_RPC_URL")
	bchTestnetUser := os.Getenv("BCASH_TESTNET_RPC_USERNAME")
	bchTestnetPassword := os.Getenv("BCASH_TESTNET_RPC_PASSWORD")
	bchTestnetNodeClient := rpc.NewClient(bchTestnetURL, bchTestnetUser, bchTestnetPassword)
	bchTestnetProxy := proxy.NewProxy(bchTestnetNodeClient)
	bchTestnetAPI := api.NewApi(btctypes.BchTestnet, bchTestnetProxy, bchCache, logger)

	// Initialize Ethereum API.
	taggedKeys := map[string]string{
		"":         os.Getenv("INFURA_KEY_DEFAULT"),
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	infuraMainnetClient := rpc.NewInfuraClient(ethtypes.EthMainnet, taggedKeys)
	ethMainnetProxy := proxy.NewProxy(infuraMainnetClient)
	ethMainnetAPI := api.NewApi(ethtypes.EthMainnet, ethMainnetProxy, ethCache, logger)

	var ethTestnetClient rpc.Client
	ethKovanRPCURL := os.Getenv("ETH_KOVAN_RPC_URL")
	if ethKovanRPCURL == "" {
		logger.Infof("Using Infura")
		ethTestnetClient = rpc.NewInfuraClient(ethtypes.EthKovan, taggedKeys)
	} else {
		logger.Infof("Using local ETH node at: %s", ethKovanRPCURL)
		ethKovanUser := os.Getenv("ETH_KOVAN_RPC_USERNAME")
		ethKovanPassword := os.Getenv("ETH_KOVAN_RPC_PASSWORD")
		ethTestnetClient = rpc.NewClient(ethKovanRPCURL, ethKovanUser, ethKovanPassword)
	}
	ethTestnetProxy := proxy.NewProxy(ethTestnetClient)
	ethTestnetAPI := api.NewApi(ethtypes.EthKovan, ethTestnetProxy, ethKovanCache, logger)

	maticTestnetClient := rpc.NewClient(os.Getenv("MATIC_TESTNET_RPC_URL"), "", "")
	maticTestnetProxy := proxy.NewProxy(maticTestnetClient)
	maticTestnetAPI := api.NewApi(ethtypes.MaticTestnet, maticTestnetProxy, maticCache, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcTestnetAPI, zecTestnetAPI, bchTestnetAPI, ethMainnetAPI, ethTestnetAPI, maticTestnetAPI)
	server.Run()
}
