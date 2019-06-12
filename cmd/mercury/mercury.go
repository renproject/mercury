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

	mainnetCSClient, err := btc.NewCS("mainnet")
	if err != nil {
		logger.Error(err)
		return
	}
	testnetCSClient, err := btc.NewCS("testnet")
	if err != nil {
		logger.Error(err)
		return
	}

	zecMainnetCSClient, err := zec.NewCS("mainnet")
	if err != nil {
		logger.Error(err)
		return
	}
	zecTestnetCSClient, err := zec.NewCS("testnet")
	if err != nil {
		logger.Error(err)
		return
	}

	// mainnetFNClient := btc.NewFN("mainnet", os.Getenv("BITCOIN_MAINNET_RPC_URL"), os.Getenv("BITCOIN_MAINNET_RPC_USER"), os.Getenv("BITCOIN_MAINNET_RPC_PASSWORD"))
	testnetFNClient := btc.NewFN("testnet", os.Getenv("BITCOIN_TESTNET_RPC_URL"), os.Getenv("TESTNET_RPC_USER"), os.Getenv("TESTNET_RPC_PASSWORD"))
	omniTestnetFNClient := btc.NewFN("testnet", os.Getenv("OMNI_TESTNET_RPC_URL"), os.Getenv("TESTNET_RPC_USER"), os.Getenv("TESTNET_RPC_PASSWORD"))
	omniMainnetFNClient := btc.NewFN("mainnet", os.Getenv("OMNI_MAINNET_RPC_URL"), os.Getenv("BITCOIN_MAINNET_RPC_USER"), os.Getenv("BITCOIN_MAINNET_RPC_PASSWORD"))

	zecTestnetFNClient := zec.NewFN("testnet", os.Getenv("ZCASH_TESTNET_RPC_URL"), os.Getenv("TESTNET_RPC_USER"), os.Getenv("TESTNET_RPC_PASSWORD"))
	// zecMainnetFNClient := zec.NewFN(os.Getenv("ZCASH_MAINNET_RPC_URL"), os.Getenv("ZCASH_MAINNET_RPC_USER"), os.Getenv("ZCASH_MAINNET_RPC_PASSWORD"), "mainnet")

	btcMainnetPlugin := btc.New("btc", btc.NewMulti(mainnetBIClient, mainnetCSClient, omniMainnetFNClient), logger)
	btcTestnetPlugin := btc.New("btc-testnet3", btc.NewMulti(testnetFNClient, testnetBIClient, testnetCSClient, omniTestnetFNClient), logger)
	zecTestnetPlugin := zec.New("zec", zec.NewMulti(zecMainnetCSClient), logger)
	zecMainnetPlugin := zec.New("zec-testnet", zec.NewMulti(zecTestnetFNClient, zecTestnetCSClient), logger)
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
