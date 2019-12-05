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

	// Initialise stores.
	store := kv.NewMemDB(kv.JSONCodec)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	btcStore := kv.NewTTLCache(ctx, store, "btc", 10*time.Second)
	btcCache := cache.New(btcStore, logger)
	zecStore := kv.NewTTLCache(ctx, store, "zec", 10*time.Second)
	zecCache := cache.New(zecStore, logger)
	bchStore := kv.NewTTLCache(ctx, store, "bch", 10*time.Second)
	bchCache := cache.New(bchStore, logger)
	ethStore := kv.NewTTLCache(ctx, store, "eth", 10*time.Second)
	ethCache := cache.New(ethStore, logger)

	// Initialise Bitcoin API.
	btcNodeClient := rpc.NewClient(os.Getenv("BTC_RPC_URL"), "user", "password")
	btcProxy := proxy.NewProxy(logger, btcNodeClient)
	btcAPI := api.NewApi(btctypes.BtcLocalnet, btcProxy, btcCache, logger)

	// Initialise ZCash API.
	zecNodeClient := rpc.NewClient(os.Getenv("ZEC_RPC_URL"), "user", "password")
	zecProxy := proxy.NewProxy(logger, zecNodeClient)
	zecAPI := api.NewApi(btctypes.ZecLocalnet, zecProxy, zecCache, logger)

	// Initialise BCash API.
	bchNodeClient := rpc.NewClient(os.Getenv("BCH_RPC_URL"), "user", "password")
	bchProxy := proxy.NewProxy(logger, bchNodeClient)
	bchAPI := api.NewApi(btctypes.BchLocalnet, bchProxy, bchCache, logger)

	ethNodeClient := rpc.NewClient(os.Getenv("ETH_RPC_URL"), "", "")
	ethProxy := proxy.NewProxy(logger, ethNodeClient)
	ethAPI := api.NewApi(ethtypes.EthLocalnet, ethProxy, ethCache, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcAPI, zecAPI, bchAPI, ethAPI)
	server.Run()
}
