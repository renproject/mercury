package main

import (
	"os"

	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialise logger.
	logger := logrus.StandardLogger()

	// Retrieve data from environment variables.
	btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USER")
	btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")

	// Initialise Bitcoin API.
	nodeClient, err := btcrpc.NewNodeClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
	if err != nil {
		logger.Fatalf("cannot construct node client: %v", err)
	}
	btcProxy := proxy.NewBtcProxy(nodeClient)
	btcBackend := api.NewBtcApi(btcProxy, logger)

	// Initialize Ethereum API.
	infuraAPIKey := os.Getenv("INFURA_KEY_DEFAULT")
	ethMainnetProxy, err := proxy.NewInfuraProxy(ethtypes.Mainnet, infuraAPIKey)
	if err != nil {
		logger.Fatalf("cannot construct infura mainnet proxy: %v", err)
	}
	ethMainnetBackend := api.NewEthBackend(ethMainnetProxy, logger)

	ethKovanProxy, err := proxy.NewInfuraProxy(ethtypes.Kovan, infuraAPIKey)
	if err != nil {
		logger.Fatalf("cannot construct infura testnet proxy: %v", err)
	}
	ethKovanBackend := api.NewEthBackend(ethKovanProxy, logger)

	// Set-up and start the server.
	server := api.NewServer(logger, "5000", btcBackend, ethMainnetBackend, ethKovanBackend)
	server.Run()
}
