package main

import (
	"context"
	"net/http"
	"os"
	"time"

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

	db := kv.NewMemDB(kv.JSONCodec)

	// Initialise stores.
	rinkebyTTL := kv.NewTTLCache(context.Background(), db, "ethRinkebyCache", 5*time.Second)
	kovanTTL := kv.NewTTLCache(context.Background(), db, "ethKovanCache", 5*time.Second)
	btcTestTTL := kv.NewTTLCache(context.Background(), db, "btcTestCache", 5*time.Second)
	zecTestTTL := kv.NewTTLCache(context.Background(), db, "zecTestCache", 5*time.Second)
	bchTestTTL := kv.NewTTLCache(context.Background(), db, "bchTestCache", 5*time.Second)
	ethTTL := kv.NewTTLCache(context.Background(), db, "ethCache", 5*time.Second)
	btcTTL := kv.NewTTLCache(context.Background(), db, "btcCache", 5*time.Second)
	zecTTL := kv.NewTTLCache(context.Background(), db, "zecCache", 5*time.Second)
	bchTTL := kv.NewTTLCache(context.Background(), db, "bchCache", 5*time.Second)

	ethRinkebyCache := cache.New(kv.NewTable(db, "ethRinkeby"), rinkebyTTL, logger)
	ethKovanCache := cache.New(kv.NewTable(db, "ethKovan"), kovanTTL, logger)
	btcTestCache := cache.New(kv.NewTable(db, "btcTest"), btcTestTTL, logger)
	zecTestCache := cache.New(kv.NewTable(db, "zecTest"), zecTestTTL, logger)
	bchTestCache := cache.New(kv.NewTable(db, "bchTest"), bchTestTTL, logger)
	ethCache := cache.New(kv.NewTable(db, "eth"), ethTTL, logger)
	btcCache := cache.New(kv.NewTable(db, "btc"), btcTTL, logger)
	zecCache := cache.New(kv.NewTable(db, "zec"), zecTTL, logger)
	bchCache := cache.New(kv.NewTable(db, "bch"), bchTTL, logger)

	// Initialise Bitcoin API.
	btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
	btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
	btcTestnetNodeClient := rpc.NewClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
	btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
	btcTestnetAPI := api.NewApi(btctypes.BtcTestnet, btcTestnetProxy, btcTestCache, logger)

	btcMainnetURL := os.Getenv("BITCOIN_MAINNET_RPC_URL")
	btcMainnetUser := os.Getenv("BITCOIN_MAINNET_RPC_USERNAME")
	btcMainnetPassword := os.Getenv("BITCOIN_MAINNET_RPC_PASSWORD")
	btcMainnetNodeClient := rpc.NewClient(btcMainnetURL, btcMainnetUser, btcMainnetPassword)
	btcMainnetProxy := proxy.NewProxy(btcMainnetNodeClient)
	btcMainnetAPI := api.NewApi(btctypes.BtcMainnet, btcMainnetProxy, btcCache, logger)

	// Initialise ZCash API.
	zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
	zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
	zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
	zecTestnetNodeClient := rpc.NewClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
	zecTestnetProxy := proxy.NewProxy(zecTestnetNodeClient)
	zecTestnetAPI := api.NewApi(btctypes.ZecTestnet, zecTestnetProxy, zecTestCache, logger)

	zecMainnetURL := os.Getenv("ZCASH_MAINNET_RPC_URL")
	zecMainnetUser := os.Getenv("ZCASH_MAINNET_RPC_USERNAME")
	zecMainnetPassword := os.Getenv("ZCASH_MAINNET_RPC_PASSWORD")
	zecMainnetNodeClient := rpc.NewClient(zecMainnetURL, zecMainnetUser, zecMainnetPassword)
	zecMainnetProxy := proxy.NewProxy(zecMainnetNodeClient)
	zecMainnetAPI := api.NewApi(btctypes.ZecMainnet, zecMainnetProxy, zecCache, logger)

	// Initialise BCash API.
	bchTestnetURL := os.Getenv("BCASH_TESTNET_RPC_URL")
	bchTestnetUser := os.Getenv("BCASH_TESTNET_RPC_USERNAME")
	bchTestnetPassword := os.Getenv("BCASH_TESTNET_RPC_PASSWORD")
	bchTestnetNodeClient := rpc.NewClient(bchTestnetURL, bchTestnetUser, bchTestnetPassword)
	bchTestnetProxy := proxy.NewProxy(bchTestnetNodeClient)
	bchTestnetAPI := api.NewApi(btctypes.BchTestnet, bchTestnetProxy, bchTestCache, logger)

	bchMainnetURL := os.Getenv("BCASH_MAINNET_RPC_URL")
	bchMainnetUser := os.Getenv("BCASH_MAINNET_RPC_USERNAME")
	bchMainnetPassword := os.Getenv("BCASH_MAINNET_RPC_PASSWORD")
	bchMainnetNodeClient := rpc.NewClient(bchMainnetURL, bchMainnetUser, bchMainnetPassword)
	bchMainnetProxy := proxy.NewProxy(bchMainnetNodeClient)
	bchMainnetAPI := api.NewApi(btctypes.BchMainnet, bchMainnetProxy, bchCache, logger)

	// Initialize Ethereum API.
	infuraKey := os.Getenv("INFURA_KEY_DEFAULT")
	client := new(http.Client)
	infuraMainnetClient := rpc.NewInfuraClient(client, ethtypes.Mainnet, infuraKey)
	ethMainnetProxy := proxy.NewProxy(infuraMainnetClient)
	ethMainnetAPI := api.NewApi(ethtypes.Mainnet, ethMainnetProxy, ethCache, logger)

	infuraRinkebyClient := rpc.NewInfuraClient(client, ethtypes.Rinkeby, infuraKey)
	ethRinkebyProxy := proxy.NewProxy(infuraRinkebyClient)
	ethRinkebyAPI := api.NewApi(ethtypes.Rinkeby, ethRinkebyProxy, ethRinkebyCache, logger)

	var testnetClient rpc.Client
	ethKovanRPCURL := os.Getenv("ETH_KOVAN_RPC_URL")
	if ethKovanRPCURL == "" {
		logger.Infof("Using Infura")
		testnetClient = rpc.NewInfuraClient(client, ethtypes.Kovan, infuraKey)
	} else {
		logger.Infof("Using local ETH node at: %s", ethKovanRPCURL)
		ethKovanUser := os.Getenv("ETH_KOVAN_RPC_USERNAME")
		ethKovanPassword := os.Getenv("ETH_KOVAN_RPC_PASSWORD")
		testnetClient = rpc.NewClient(ethKovanRPCURL, ethKovanUser, ethKovanPassword)
	}
	ethTestnetProxy := proxy.NewProxy(testnetClient)
	ethTestnetAPI := api.NewApi(ethtypes.Kovan, ethTestnetProxy, ethKovanCache, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcMainnetAPI, zecMainnetAPI, bchMainnetAPI, btcTestnetAPI, zecTestnetAPI, bchTestnetAPI, ethMainnetAPI, ethTestnetAPI, ethRinkebyAPI)
	server.Run()
}
