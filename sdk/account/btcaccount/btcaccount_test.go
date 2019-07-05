package btcaccount_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/account/btcaccount"

	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc account ", func() {
	Context("when fetching utxos", func() {
		It("should fetch at least one utxo from the funded account", func() {
			// Get the account with actual balance
			client := btcclient.NewBtcClient(btctypes.Testnet)
			wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := New(logrus.StandardLogger(), client, key)
			Expect(err).NotTo(HaveOccurred())
			utxos, err := account.UTXOs(context.Background(), btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should fetch zero utxos from a random account", func() {
			client := btcclient.NewBtcClient(btctypes.Testnet)
			account, err := RandomAccount(logrus.StandardLogger(), client)
			Expect(err).NotTo(HaveOccurred())
			utxos, err := account.UTXOs(context.Background(), btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
			Expect(err).NotTo(HaveOccurred())
			// fmt.Printf("address: %v has balance: %v\n", account.Address().EncodeAddress(), balance)
			Expect(len(utxos)).Should(Equal(0))
		})
	})

	Context("when transferring funds ", func() {
		It("should be able to transfer funds to itself", func() {
			// Get the account with actual balance
			client := btcclient.NewBtcClient(btctypes.Testnet)
			wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := New(logrus.StandardLogger(), client, key)
			Expect(err).NotTo(HaveOccurred())
			utxos, err := account.UTXOs(context.Background(), btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress := account.Address()
			Expect(err).NotTo(HaveOccurred())
			amount := 180000 * btctypes.SAT
			fee := 10000 * btctypes.SAT
			balance := utxos.Sum()
			Expect(balance).Should(BeNumerically(">=", amount+fee))
			err = account.Transfer(context.Background(), toAddress, amount, fee)
			Expect(err).NotTo(HaveOccurred())

			// the following tests aren't going to work accurately because of the node indexing
			newUTXOs, err := account.UTXOs(context.Background(), btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
			Expect(err).NotTo(HaveOccurred())

			// Our original account should have less balance
			sourceBalance := newUTXOs.Sum()
			Expect(sourceBalance).Should(Equal(balance - fee))
		})
	})
})
