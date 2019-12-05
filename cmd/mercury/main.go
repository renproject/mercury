package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := kv.NewMemDB(kv.JSONCodec)

	// Initialise stores.
	ethRinkebyCache := cache.New(kv.NewTTLCache(ctx, db, "ethRinkeby", 10*time.Second), logger)
	ethKovanCache := cache.New(kv.NewTTLCache(ctx, db, "ethKovan", 10*time.Second), logger)
	btcTestCache := cache.New(kv.NewTTLCache(ctx, db, "btcTest", 10*time.Second), logger)
	zecTestCache := cache.New(kv.NewTTLCache(ctx, db, "zecTest", 10*time.Second), logger)
	bchTestCache := cache.New(kv.NewTTLCache(ctx, db, "bchTest", 10*time.Second), logger)
	ethCache := cache.New(kv.NewTTLCache(ctx, db, "eth", 10*time.Second), logger)
	btcCache := cache.New(kv.NewTTLCache(ctx, db, "btc", 10*time.Second), logger)
	zecCache := cache.New(kv.NewTTLCache(ctx, db, "zec", 10*time.Second), logger)
	bchCache := cache.New(kv.NewTTLCache(ctx, db, "bch", 10*time.Second), logger)

	// Initialise Bitcoin API.
	btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
	btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
	btcTestnetNodeClient := rpc.NewClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
	btcTestnetProxy := proxy.NewProxy(logger, btcTestnetNodeClient)
	btcTestnetAPI := api.NewApi(btctypes.BtcTestnet, btcTestnetProxy, btcTestCache, logger)

	btcMainnetURL := os.Getenv("BITCOIN_MAINNET_RPC_URL")
	btcMainnetUser := os.Getenv("BITCOIN_MAINNET_RPC_USERNAME")
	btcMainnetPassword := os.Getenv("BITCOIN_MAINNET_RPC_PASSWORD")
	btcMainnetNodeClient := rpc.NewClient(btcMainnetURL, btcMainnetUser, btcMainnetPassword)
	btcMainnetProxy := proxy.NewProxy(logger, btcMainnetNodeClient)
	btcMainnetAPI := api.NewApi(btctypes.BtcMainnet, btcMainnetProxy, btcCache, logger)

	// Initialise ZCash API.
	zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
	zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
	zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
	zecTestnetNodeClient := rpc.NewClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
	zecTestnetProxy := proxy.NewProxy(logger, zecTestnetNodeClient)
	zecTestnetAPI := api.NewApi(btctypes.ZecTestnet, zecTestnetProxy, zecTestCache, logger)

	zecMainnetURL := os.Getenv("ZCASH_MAINNET_RPC_URL")
	zecMainnetUser := os.Getenv("ZCASH_MAINNET_RPC_USERNAME")
	zecMainnetPassword := os.Getenv("ZCASH_MAINNET_RPC_PASSWORD")
	zecMainnetNodeClient := rpc.NewClient(zecMainnetURL, zecMainnetUser, zecMainnetPassword)
	zecMainnetProxy := proxy.NewProxy(logger, zecMainnetNodeClient)
	zecMainnetAPI := api.NewApi(btctypes.ZecMainnet, zecMainnetProxy, zecCache, logger)

	// Initialise BCash API.
	bchTestnetURL := os.Getenv("BCASH_TESTNET_RPC_URL")
	bchTestnetUser := os.Getenv("BCASH_TESTNET_RPC_USERNAME")
	bchTestnetPassword := os.Getenv("BCASH_TESTNET_RPC_PASSWORD")
	bchTestnetNodeClient := rpc.NewClient(bchTestnetURL, bchTestnetUser, bchTestnetPassword)
	bchTestnetProxy := proxy.NewProxy(logger, bchTestnetNodeClient)
	bchTestnetAPI := api.NewApi(btctypes.BchTestnet, bchTestnetProxy, bchTestCache, logger)

	bchMainnetURL := os.Getenv("BCASH_MAINNET_RPC_URL")
	bchMainnetUser := os.Getenv("BCASH_MAINNET_RPC_USERNAME")
	bchMainnetPassword := os.Getenv("BCASH_MAINNET_RPC_PASSWORD")
	bchMainnetNodeClient := rpc.NewClient(bchMainnetURL, bchMainnetUser, bchMainnetPassword)
	bchMainnetProxy := proxy.NewProxy(logger, bchMainnetNodeClient)
	bchMainnetAPI := api.NewApi(btctypes.BchMainnet, bchMainnetProxy, bchCache, logger)

	// Initialize Ethereum API.
	taggedKeys := map[string]string{
		"":         os.Getenv("INFURA_KEY_DEFAULT"),
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	infuraMainnetClient := rpc.NewInfuraClient(ethtypes.Mainnet, taggedKeys)
	ethMainnetProxy := proxy.NewProxy(logger, infuraMainnetClient)
	ethMainnetAPI := api.NewApi(ethtypes.Mainnet, ethMainnetProxy, ethCache, logger)

	infuraRinkebyClient := rpc.NewInfuraClient(ethtypes.Rinkeby, taggedKeys)
	ethRinkebyProxy := proxy.NewProxy(logger, infuraRinkebyClient)
	ethRinkebyAPI := api.NewApi(ethtypes.Rinkeby, ethRinkebyProxy, ethRinkebyCache, logger)

	var testnetClient rpc.Client
	ethKovanRPCURL := os.Getenv("ETH_KOVAN_RPC_URL")
	if ethKovanRPCURL == "" {
		logger.Infof("Using Infura")
		testnetClient = rpc.NewInfuraClient(ethtypes.Kovan, taggedKeys)
	} else {
		logger.Infof("Using local ETH node at: %s", ethKovanRPCURL)
		ethKovanUser := os.Getenv("ETH_KOVAN_RPC_USERNAME")
		ethKovanPassword := os.Getenv("ETH_KOVAN_RPC_PASSWORD")
		testnetClient = rpc.NewClient(ethKovanRPCURL, ethKovanUser, ethKovanPassword)
	}
	ethTestnetProxy := proxy.NewProxy(logger, testnetClient)
	ethTestnetAPI := api.NewApi(ethtypes.Kovan, ethTestnetProxy, ethKovanCache, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcMainnetAPI, zecMainnetAPI, bchMainnetAPI, btcTestnetAPI, zecTestnetAPI, bchTestnetAPI, ethMainnetAPI, ethTestnetAPI, ethRinkebyAPI)
	server.Run()
}
