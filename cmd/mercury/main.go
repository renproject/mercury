package main

import (
	"os"

	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up logger
	logger := logrus.StandardLogger()

	// Getting data from env variables
	btcHost := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcUser := os.Getenv("BITCOIN_TESTNET_RPC_USER")
	btcPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")

	// Initialize BTC proxy
	nodeClient, err := btcrpc.NewNodeClient(btctypes.Testnet, btcHost, btcUser, btcPassword)
	if err != nil {
		panic(err)
	}
	btcProxy := proxy.NewBtcProxy(btctypes.Testnet, nodeClient, btcrpc.NewChainsoClient(btctypes.Testnet))
	btcBackend := api.NewBtcBackend(btcProxy)

	// Initialize ETH proxy
	ethProxy, err := proxy.NewEthProxy(ethtypes.EthKovan)
	if err != nil {
		panic(err)
	}
	apiKeys := map[string]string{
		"":         os.Getenv("INFURA_KEY_DEFAULT"),
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	ethBackend := api.NewEthBackend(ethProxy, apiKeys, logger)

	// Set up the server and start running
	server := api.NewServer(logger, "5000", btcBackend, ethBackend)
	server.Run()
}
