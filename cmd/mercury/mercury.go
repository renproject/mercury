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
			logrus.WarnLevel,
		})
		hook.Timeout = 5 * time.Second
		if err != nil {
			logrus.Fatalf("cannot create a sentry hook: %v", err)
		}
		logger.AddHook(hook)
	}

	testnetFNClient := btc.NewFN("testnet", os.Getenv("BITCOIN_TESTNET_RPC_URL"), os.Getenv("BITCOIN_TESTNET_RPC_USERNAME"), os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD"))
	zecTestnetFNClient := zec.NewFN("testnet", os.Getenv("ZCASH_TESTNET_RPC_URL"), os.Getenv("ZCASH_TESTNET_RPC_USERNAME"), os.Getenv("ZCASH_TESTNET_RPC_PASSWORD"))
	btcTestnetPlugin := btc.New("btc-testnet3", testnetFNClient, logger)
	zecTestnetPlugin := zec.New("zec-testnet", zecTestnetFNClient, logger)
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
	mercury.New(os.Getenv("PORT"), logger, btcTestnetPlugin, zecTestnetPlugin, kovanEthPlugin, ropstenEthPlugin, mainnetEthPlugin).Run()
}
