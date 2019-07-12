package btcclient_test

import (
	"encoding/hex"
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
	logger := logrus.StandardLogger()

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutils.HdKey {
		wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
	}

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

	Context("when fetching UTXOs", func() {
		It("should return the UTXO for a transaction with unspent outputs", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			txHash := btctypes.TxHash("bd4bb310b0c6c4e5225bc60711931552e5227c94ef7569bfc7037f014d91030c")
			index := uint32(0)
			utxo, err := client.UTXO(txHash, index)
			Expect(err).NotTo(HaveOccurred())
			Expect(utxo.TxHash).To(Equal(txHash))
			Expect(utxo.Amount).To(Equal(btctypes.Amount(100000)))
			Expect(utxo.ScriptPubKey).To(Equal("76a9142d2b683141de54613e7c6648afdb454fa3b4126d88ac"))
			Expect(utxo.Vout).To(Equal(index))
		})

		It("should return an error for an invalid UTXO index", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.UTXO("bd4bb310b0c6c4e5225bc60711931552e5227c94ef7569bfc7037f014d91030c", 3)
			Expect(err).To(Equal(ErrUTXOSpent))
		})

		It("should return an error for a UTXO that has been spent", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.UTXO("7e65d34373491653934d32cc992211b14b9e0e80d4bb9380e97aaa05fa872df5", 0)
			Expect(err).To(Equal(ErrUTXOSpent))
		})

		It("should return an error for an invalid transaction hash", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.UTXO("abcdefg", 0)
			Expect(err).To(Equal(ErrInvalidTxHash))
		})

		It("should return an error for a non-existent transaction hash", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.UTXO("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1", 0)
			Expect(err).To(Equal(ErrTxHashNotFound))
		})

		It("should return a non-zero number of UTXOs for a funded address that has been imported", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := loadTestAccounts(btctypes.Localnet).BTCAddress(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOsFromAddress(address)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(BeNumerically(">", 0))
		})

		It("should return zero UTXOs for a randomly generated address", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := testutils.RandomBTCAddress(btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.UTXOsFromAddress(address)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(Equal(0))
		})
	})

	tx := func(client Client, address btctypes.Address) btctypes.StandardTx {
		recipient, err := loadTestAccounts(btctypes.Localnet).BTCAddress(44, 1, 0, 0, 2)
		Expect(err).NotTo(HaveOccurred())

		utxos := btctypes.UTXOs{
			{
				TxHash: "4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1",
				Amount: 20000,
				Vout:   0,
			},
			{
				TxHash: "1b112299cc23d43935619437343d56e75fd278af50aaa79e0839bb4d08e0e9b4",
				Amount: 80000,
				Vout:   1,
			},
		}
		recipients := btctypes.Recipients{
			{
				Address: recipient,
				Amount:  40000,
			},
		}

		tx, err := client.BuildUnsignedTx(utxos, recipients, address, 600)
		Expect(err).ToNot(HaveOccurred())

		return tx
	}

	Context("when building a utx", func() {
		It("should return the expected serialized transaction", func() {
			client, err := New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			address, err := loadTestAccounts(btctypes.Localnet).BTCAddress(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())

			// Build unsigned transaction.
			tx := tx(client, address)
			serializedTx := hex.EncodeToString(tx.Serialize())

			// Validate serialized transaction. We expect it to be one of the following two due to some non-determinism
			// with Golang maps in the recipient ordering. This can be removed if we stop relying on the
			// `rpcclient.CreateRawTransaction` function.
			expectedTxs := []string{
				"0200000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff0208e80000000000001976a9142b075b01d5dd314a902357617ed67f16e4e0591388ac409c0000000000001976a914a4cfcb06f8f41446c9142a2c1f494014563aafd788ac00000000",
				"0200000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a914a4cfcb06f8f41446c9142a2c1f494014563aafd788ac08e80000000000001976a9142b075b01d5dd314a902357617ed67f16e4e0591388ac00000000",
			}
			Expect(expectedTxs).To(ContainElement(serializedTx))
		})
	})
})