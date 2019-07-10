package zecclient_test

import (
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/zecclient"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/zecrpc"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/zectypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("zec client", func() {
	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network zectypes.Network) testutils.HdKey {
		wallet, err := testutils.LoadHdWalletFromEnv("ZEC_TEST_MNEMONIC", "ZEC_TEST_PASSPHRASE", network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
	}

	BeforeSuite(func() {
		logger := logrus.StandardLogger()
		store := kv.NewJSON(kv.NewMemDB())
		cache := cache.New(store, logger)

		zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
		zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
		zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
		zecTestnetNodeClient, err := zecrpc.NewNodeClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
		Expect(err).ToNot(HaveOccurred())

		zecTestnetProxy := proxy.NewProxy(zecTestnetNodeClient)
		zecTestnetAPI := api.NewZecApi(zectypes.Testnet, zecTestnetProxy, cache, logger)
		server := api.NewServer(logger, "5000", zecTestnetAPI)
		go server.Run()
	})

	Context("when fetching UTXOs", func() {
		It("should return a non-zero number of UTXOs for a transaction with unspent outputs", func() {
			client, err := New(zectypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOs("e96953b5030f44686e71650d6cb71a83625059ad086f7fc7802775e22cef0f65")
			Expect(err).NotTo(HaveOccurred())
			log.Println(utxos)
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should return zero UTXOs for a transaction with spent outputs", func() {
			client, err := New(zectypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			address, err := loadTestAccounts(zectypes.Localnet).ZECAddress(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())

			log.Println(address)

			utxos, err := client.UTXOs("d16e32d5e7b5442c8aaffe687ed0db7c2b4a7221a8607620902c06b214f8c4b1")
			log.Println(utxos)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(Equal(0))
		})

		It("should return an error for an invalid transaction hash", func() {
			client, err := New(zectypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOs("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1")
			Expect(err).To(HaveOccurred())
			Expect(len(utxos)).Should(Equal(0))
		})

		It("should return a non-zero number of UTXOs for a funded address that has been imported", func() {
			client, err := New(zectypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := loadTestAccounts(zectypes.Localnet).ZECAddress(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOsFromAddress(address)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should return zero UTXOs for a randomly generated address", func() {
			client, err := New(zectypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := testutils.RandomZECAddress(zectypes.Localnet)
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
