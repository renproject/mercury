package btcaccount_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/testutil/btcaccount"

	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc account", func() {
	logger := logrus.StandardLogger()

	Context("when fetching utxos", func() {
		It("should fetch at least one utxo from the funded account", func() {
			// Get the account with actual balance
			client := btcclient.NewClient(logger, btctypes.BtcLocalnet)
			wallet, err := testutil.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should fetch zero utxos from a random account", func() {
			client := btcclient.NewClient(logger, btctypes.BtcLocalnet)
			account, err := RandomAccount(client)
			Expect(err).NotTo(HaveOccurred())
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			// fmt.Printf("address: %v has balance: %v\n", account.Address().EncodeAddress(), balance)
			Expect(len(utxos)).Should(Equal(0))
		})
	})

	// FIXME: Do not run multiple tests with the same keypair.
	Context("when transferring funds", func() {
		It("should be able to transfer funds to itself", func() {
			// Get the account with actual balance
			client := btcclient.NewClient(logger, btctypes.BtcLocalnet)
			wallet, err := testutil.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 2)
			Expect(err).NotTo(HaveOccurred())
			account, err := NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("from address: ", account.Address().EncodeAddress())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress := account.Address()
			amount := 50000 * btctypes.SAT
			fmt.Println("to address: ", toAddress.EncodeAddress())

			txHash, err := account.Transfer(ctx, toAddress, amount, types.Standard, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("txHash: ", txHash[:])
		})

		It("should be able to transfer funds to itself (legacy address)", func() {
			// Get the account with actual balance
			client := btcclient.NewClient(logger, btctypes.BchLocalnet)
			wallet, err := testutil.LoadHdWalletFromEnv("BCH_TEST_MNEMONIC", "BCH_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 2)
			Expect(err).NotTo(HaveOccurred())
			account, err := NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("from address: ", account.Address().EncodeAddress())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress, err := btctypes.AddressFromPubKey(key.PublicKey, btctypes.BtcLocalnet)
			Expect(err).NotTo(HaveOccurred())
			amount := 50000 * btctypes.SAT
			fmt.Println("to address: ", toAddress.EncodeAddress())

			txHash, err := account.Transfer(ctx, toAddress, amount, types.Standard, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("txHash: ", txHash[:])
		})

		It("should be able to transfer funds to itself (cash address)", func() {
			// Get the account with actual balance
			client := btcclient.NewClient(logger, btctypes.BchLocalnet)
			wallet, err := testutil.LoadHdWalletFromEnv("BCH_TEST_MNEMONIC", "BCH_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 2)
			Expect(err).NotTo(HaveOccurred())
			account, err := NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("from address: ", account.Address().EncodeAddress())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress := account.Address()
			amount := 50000 * btctypes.SAT
			fmt.Println("to address: ", toAddress.EncodeAddress())

			txHash, err := account.Transfer(ctx, toAddress, amount, types.Standard, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("txHash: ", txHash[:])
		})

		It("should be able to transfer funds to itself", func() {
			// Get the account with actual balance
			client := btcclient.NewClient(logger, btctypes.ZecLocalnet)
			wallet, err := testutil.LoadHdWalletFromEnv("ZEC_TEST_MNEMONIC", "ZEC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account, err := NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("from address: ", account.Address().EncodeAddress())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress := account.Address()
			amount := 50000 * btctypes.SAT
			fmt.Println("to address: ", toAddress.EncodeAddress())

			txHash, err := account.Transfer(ctx, toAddress, amount, types.Standard, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("txHash: ", txHash[:])
		})

		It("should be able to transfer funds to itself using SegWit", func() {
			// Get the account with actual balance
			client := btcclient.NewClient(logger, btctypes.BtcLocalnet)
			wallet, err := testutil.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", client.Network())
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 2)
			Expect(err).NotTo(HaveOccurred())
			account, err := NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("from address: ", account.Address().EncodeAddress())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			utxos, err := account.UTXOs(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))

			// Build the transaction
			toAddress, err := btctypes.SegWitAddressFromPubKey(key.PublicKey, btctypes.BtcLocalnet)
			Expect(err).NotTo(HaveOccurred())

			amount := 50000 * btctypes.SAT
			fmt.Println("to address: ", toAddress.EncodeAddress())

			txHash, err := account.Transfer(ctx, toAddress, amount, types.Standard, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("txHash: ", txHash[:])
		})
	})
})
