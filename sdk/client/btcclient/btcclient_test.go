package btcclient_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc client", func() {
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

	Context("when fetching UTXOs", func() {
		It("should return a non-zero number of UTXOs for a transaction with unspent outputs", func() {
			client, err := New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOs("bd4bb310b0c6c4e5225bc60711931552e5227c94ef7569bfc7037f014d91030c")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should return zero UTXOs for a transaction with spent outputs", func() {
			client, err := New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOs("7e65d34373491653934d32cc992211b14b9e0e80d4bb9380e97aaa05fa872df5")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(Equal(0))
		})

		It("should return an error for an invalid transaction hash", func() {
			client, err := New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOs("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1")
			Expect(err).To(HaveOccurred())
			Expect(len(utxos)).Should(Equal(0))
		})

		It("should return a non-zero number of UTXOs for a funded address that has been imported", func() {
			client, err := New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := loadTestAccounts(btctypes.Localnet).Address(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOsFromAddress(address)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should return zero UTXOs for a randomly generated address", func() {
			client, err := New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := testutils.RandomAddress(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOsFromAddress(address)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(Equal(0))
		})
	})

	Context("when building a utx", func() {
		PIt("should have the correct inputs", func() {
			// TODO: write the test
		})

		PIt("should have the correct outputs", func() {
			// TODO: write the test
		})
	})

	Context("when submitting stx to bitcoin", func() {
		PIt("should be able to submit a stx", func() {
			// TODO: write the test
		})
	})
})
