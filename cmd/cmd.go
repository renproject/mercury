package main

import (
	"os"

	"github.com/renproject/mercury"
	"github.com/renproject/mercury/btc"
	"github.com/renproject/mercury/eth"
)

func main() {
	btcTestnetPlugin, err := btc.New(os.Getenv("BITCOIN_TESTNET_RPC_URL"), os.Getenv("BITCOIN_TESTNET_RPC_USER"), os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD"))
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

	kovanEthPlugin := eth.New("kovan", apiKeys)
	if err != nil {
		panic(err)
	}
	ropstenEthPlugin := eth.New("ropsten", apiKeys)
	if err != nil {
		panic(err)
	}
	mainnetEthPlugin := eth.New("mainnet", apiKeys)
	if err != nil {
		panic(err)
	}
	mercury.New("8123", btcTestnetPlugin, kovanEthPlugin, ropstenEthPlugin, mainnetEthPlugin).Run()
}
