package main

import (
	"os"

	"github.com/renproject/mercury"
	"github.com/renproject/mercury/btc"
	"github.com/renproject/mercury/eth"
	"github.com/renproject/mercury/zec"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.StandardLogger()

	mainnetClient, err := btc.NewBI("mainnet")
	if err != nil {
		logger.Error(err)
		return
	}

	testnetClient, err := btc.NewBI("testnet")
	if err != nil {
		logger.Error(err)
		return
	}

	btcMainnetPlugin := btc.New("btc", mainnetClient)
	btcTestnetPlugin := btc.New("btc-testnet3", testnetClient)
	zecTestnetPlugin := zec.New("zec-testnet", os.Getenv("ZCASH_TESTNET_RPC_URL"), os.Getenv("ZCASH_TESTNET_RPC_USER"), os.Getenv("ZCASH_TESTNET_RPC_PASSWORD"))
	zecMainnetPlugin := zec.New("zec", os.Getenv("ZCASH_MAINNET_RPC_URL"), os.Getenv("ZCASH_MAINNET_RPC_USER"), os.Getenv("ZCASH_MAINNET_RPC_PASSWORD"))
	apiKeys := map[string]string{
		"":         os.Getenv("INFURA_KEY_DEFAULT"),
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	privKey := os.Getenv("ETHEREUM_PRIVATE_KEY")
	kovanEthPlugin := eth.New("eth-kovan", privKey, apiKeys)
	ropstenEthPlugin := eth.New("eth-ropsten", privKey, apiKeys)
	mainnetEthPlugin := eth.New("eth", privKey, apiKeys)
	mercury.New(os.Getenv("PORT"), logger, btcMainnetPlugin, zecMainnetPlugin,
		btcTestnetPlugin, zecTestnetPlugin, kovanEthPlugin, ropstenEthPlugin,
		mainnetEthPlugin).Run()
}
