package main

import (
	"os"
	"time"

	"github.com/evalphobia/logrus_sentry"
	"github.com/getsentry/raven-go"
	"github.com/renproject/mercury"
	"github.com/renproject/mercury/btc"
	"github.com/renproject/mercury/eth"
	"github.com/renproject/mercury/zec"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.StandardLogger()
	if sentryDSN := os.Getenv("SENTRY_DSN"); sentryDSN != "" {
		client, err := raven.New(sentryDSN)
		if err != nil {
			logrus.Fatalf("cannot connect to sentry: %v", err)
		}
		hook, err := logrus_sentry.NewWithClientSentryHook(client, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})
		hook.Timeout = 5 * time.Second
		if err != nil {
			logrus.Fatalf("cannot create a sentry hook: %v", err)
		}
		logger.AddHook(hook)
	}

	mainnetBIClient, err := btc.NewBI("mainnet")
	if err != nil {
		logger.Error(err)
		return
	}
	testnetBIClient, err := btc.NewBI("testnet")
	if err != nil {
		logger.Error(err)
		return
	}
	// mainnetFNClient := btc.NewFN("mainnet", os.Getenv("BITCOIN_MAINNET_RPC_URL"), os.Getenv("BITCOIN_MAINNET_RPC_USER"), os.Getenv("BITCOIN_MAINNET_RPC_PASSWORD"))
	testnetFNClient := btc.NewFN("testnet", os.Getenv("BITCOIN_TESTNET_RPC_URL"), os.Getenv("BITCOIN_TESTNET_RPC_USER"), os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD"))

	btcMainnetPlugin := btc.New("btc", btc.NewMulti(mainnetBIClient), logger)
	btcTestnetPlugin := btc.New("btc-testnet3", btc.NewMulti(testnetBIClient, testnetFNClient), logger)
	zecTestnetPlugin := zec.New("zec-testnet", os.Getenv("ZCASH_TESTNET_RPC_URL"), os.Getenv("ZCASH_TESTNET_RPC_USER"), os.Getenv("ZCASH_TESTNET_RPC_PASSWORD"), logger)
	zecMainnetPlugin := zec.New("zec", os.Getenv("ZCASH_MAINNET_RPC_URL"), os.Getenv("ZCASH_MAINNET_RPC_USER"), os.Getenv("ZCASH_MAINNET_RPC_PASSWORD"), logger)
	apiKeys := map[string]string{
		"":         os.Getenv("INFURA_KEY_DEFAULT"),
		"swapperd": os.Getenv("INFURA_KEY_SWAPPERD"),
		"darknode": os.Getenv("INFURA_KEY_DARKNODE"),
		"renex":    os.Getenv("INFURA_KEY_RENEX"),
		"renex-ui": os.Getenv("INFURA_KEY_RENEX_UI"),
		"dcc":      os.Getenv("INFURA_KEY_DCC"),
	}
	privKey := os.Getenv("ETHEREUM_PRIVATE_KEY")
	kovanEthPlugin := eth.New("eth-kovan", privKey, apiKeys, logger)
	ropstenEthPlugin := eth.New("eth-ropsten", privKey, apiKeys, logger)
	mainnetEthPlugin := eth.New("eth", privKey, apiKeys, logger)
	mercury.New(os.Getenv("PORT"), logger, btcMainnetPlugin, zecMainnetPlugin,
		btcTestnetPlugin, zecTestnetPlugin, kovanEthPlugin, ropstenEthPlugin,
		mainnetEthPlugin).Run()
}
