package btcgateway_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/gateway/btcgateway"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc gateway", func() {
	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutils.HdKey {
		wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
	}

	BeforeSuite(func() {
		logger := logrus.StandardLogger()
		store := kv.NewJSON(kv.NewMemDB())
		cache := cache.New(store, logger)

		btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
		btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
		btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
		btcTestnetNodeClient, err := btcrpc.NewNodeClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
		Expect(err).ToNot(HaveOccurred())

		btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
		btcTestnetAPI := api.NewBtcApi(btctypes.Testnet, btcTestnetProxy, cache, logger)
		server := api.NewServer(logger, "5000", btcTestnetAPI)
		go server.Run()
	})

	Context("when generating gateways", func() {
		PIt("should be able to generate a gateway", func() {
			client, err := btcclient.NewBtcClient(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			key, err := loadTestAccounts(btctypes.Localnet).EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			_ = New(client, &key.PublicKey, []byte{})
		})
	})
})
