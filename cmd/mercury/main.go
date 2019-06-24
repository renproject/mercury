package main

import (
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up logger
	logger := logrus.StandardLogger()

	// Getting data from env variables
	// btcHost := os.Getenv("BITCOIN_TESTNET_RPC_URL")
	// btcUser := os.Getenv("BITCOIN_TESTNET_RPC_USER")
	// btcPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
	//
	// // Initialize BTC proxy
	// nodeClient, err := btcrpc.NewNodeClient(btctypes.Testnet, "abc",btcUser,btcPassword)
	// if err != nil {
	// 	panic(err)
	// }
	btcProxy := proxy.NewBtcProxy(btctypes.Testnet, btcrpc.NewChainsoClient(btctypes.Testnet) )
	btcBackend := api.NewBtcBackend(btcProxy)

	// Set up the server and start running
	server := api.NewServer(logger, "5000", btcBackend)
	server.Run()
}
