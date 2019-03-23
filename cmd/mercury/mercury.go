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
	// btcMainnetPlugin := btc.New("btc", os.Getenv("BITCOIN_MAINNET_RPC_URL"), os.Getenv("BITCOIN_MAINNET_RPC_USER"), os.Getenv("BITCOIN_MAINNET_RPC_PASSWORD"))
	btcTestnetPlugin := btc.New("btc-testnet3", os.Getenv("BITCOIN_TESTNET_RPC_URL"), os.Getenv("BITCOIN_TESTNET_RPC_USER"), os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD"))
	zecTestnetPlugin := zec.New("zec-testnet", os.Getenv("ZCASH_TESTNET_RPC_URL"), os.Getenv("ZCASH_TESTNET_RPC_USER"), os.Getenv("ZCASH_TESTNET_RPC_PASSWORD"))
	apiKeys := map[string]string{
		"":         os.Getenv("INFURA_KEY_DEFAULT"),
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	kovanEthPlugin := eth.New("kovan", apiKeys)
	ropstenEthPlugin := eth.New("ropsten", apiKeys)
	mainnetEthPlugin := eth.New("mainnet", apiKeys)
	mercury.New(os.Getenv("PORT"), logger /*btcMainnetPlugin,*/, btcTestnetPlugin, zecTestnetPlugin, kovanEthPlugin, ropstenEthPlugin, mainnetEthPlugin).Run()
}
