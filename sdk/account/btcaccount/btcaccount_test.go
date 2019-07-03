package btcaccount_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/account/btcaccount"

	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc account ", func() {
	Context("when generating a random account", func() {
		It("should have a zero balance", func() {
			client := btcclient.NewBtcClient(btctypes.Testnet)
			account, err := RandomAccount(logrus.StandardLogger(), client)
			Expect(err).NotTo(HaveOccurred())
			balance, err := account.Balance(context.Background())
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("address: %v has balance: %v\n", account.Address().EncodeAddress(), balance)
			Expect(balance).Should(Equal(btctypes.Amount(0)))
		})
	})

	Context("when transferring funds ", func() {
		It("should build the correct transaction and broadcast it", func() {
			// Get the account with actual balance
			client := btcclient.NewBtcClient(btctypes.Testnet)
			wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE")
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := New(logrus.StandardLogger(), client, key)
			Expect(err).NotTo(HaveOccurred())
			balance, err := account.Balance(context.Background())
			Expect(err).NotTo(HaveOccurred())
			// Ensure the balance is actual positive before we transfer
			fmt.Printf("address: %v has balance: %v\n", account.Address().EncodeAddress(), balance)
			Expect(balance > 0).Should(BeTrue())

			// Create a random account to receive the funds
			toAccount, err := RandomAccount(logrus.StandardLogger(), client)
			Expect(err).NotTo(HaveOccurred())
			amount := 180000 * btctypes.SAT
			fee := 10000 * btctypes.SAT
			err = account.Transfer(context.Background(), toAccount.Address(), amount, fee)
			Expect(err).NotTo(HaveOccurred())
			newBalance, err := toAccount.Balance(context.Background())
			Expect(err).NotTo(HaveOccurred())

			// The random account should now have the correct balance
			Expect(newBalance).Should(Equal(amount))

			// Our original account should have less balance
			sourceBalance, err := account.Balance(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(sourceBalance).Should(Equal(balance - amount - fee))
		})
	})
})
