package testbtc_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/testutils/testbtc"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc account ", func() {
	logger := logrus.StandardLogger()

	BeforeSuite(func() {
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

	Context("when fetching utxos", func() {
		It("should fetch at least one utxo from the funded account", func() {
			// Get the account with actual balance
			client, err := btcclient.New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := New(client, key)
			Expect(err).NotTo(HaveOccurred())
			utxos, err := account.UTXOs()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should fetch zero utxos from a random account", func() {
			client, err := btcclient.New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			account, err := RandomAccount(client)
			Expect(err).NotTo(HaveOccurred())
			utxos, err := account.UTXOs()
			Expect(err).NotTo(HaveOccurred())
			// fmt.Printf("address: %v has balance: %v\n", account.Address().EncodeAddress(), balance)
			Expect(len(utxos)).Should(Equal(0))
		})
	})

	Context("when transferring funds ", func() {
		It("should be able to transfer funds to itself", func() {
			// Get the account with actual balance
			client, err := btcclient.New(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := New(client, key)
			Expect(err).NotTo(HaveOccurred())
			utxos, err := account.UTXOs()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress := account.Address()
			Expect(err).NotTo(HaveOccurred())
			amount := 20000 * btctypes.SAT

			_, err = account.Transfer(toAddress, amount, types.Standard)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})