package btcclient_test

import (
	"encoding/hex"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types"

	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btctx"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc client", func() {
	logger := logrus.StandardLogger()

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutil.HdKey {
		switch network.Chain() {
		case types.Bitcoin:
			wallet, err := testutil.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", network)
			Expect(err).NotTo(HaveOccurred())
			return wallet
		case types.ZCash:
			wallet, err := testutil.LoadHdWalletFromEnv("ZEC_TEST_MNEMONIC", "ZEC_TEST_PASSPHRASE", network)
			Expect(err).NotTo(HaveOccurred())
			return wallet
		default:
			panic(types.ErrUnknownChain)
		}
	}

	testCases := []struct {
		Network btctypes.Network

		// Unspent UTXO
		UnspentTxHash types.TxHash
		UnspentVout   uint32

		// Spent UTXO
		SpentTxHash types.TxHash
		SpentVout   uint32

		NonExistentTxHash types.TxHash

		ScriptPubKey string
		Amount       btctypes.Amount

		ExpectedTx string
	}{
		{
			btctypes.BtcTestnet,
			"bd4bb310b0c6c4e5225bc60711931552e5227c94ef7569bfc7037f014d91030c",
			0,
			"7e65d34373491653934d32cc992211b14b9e0e80d4bb9380e97aaa05fa872df5",
			0,
			"4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1",
			"76a9142d2b683141de54613e7c6648afdb454fa3b4126d88ac",
			100000,
			"0200000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff0208e80000000000001976a9142b075b01d5dd314a902357617ed67f16e4e0591388ac409c0000000000001976a914a4cfcb06f8f41446c9142a2c1f494014563aafd788ac00000000",
		},
		{
			btctypes.ZecTestnet,
			"41ec71582bc44fb9abc2c5d2009d1352e7df118def521b3b17c5bff86e5cfb46",
			1,
			"e96953b5030f44686e71650d6cb71a83625059ad086f7fc7802775e22cef0f65",
			0,
			"4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1",
			"76a9143735df7c4d831491ce9dc462e6f606f6faffb5ca88ac",
			100000000,
			"0400008085202f8902b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff0208e80000000000001976a914d125189e1002f3f1c948e2e123dc2926db2efb5188ac409c0000000000001976a9143735df7c4d831491ce9dc462e6f606f6faffb5ca88ac00000000809698000000000000000000000000",
		},
	}

	tx := func(client Client, address btcaddress.Address) btctx.BtcTx {
		recipient, err := loadTestAccounts(client.Network()).Address(44, 1, 0, 0, 2)
		Expect(err).NotTo(HaveOccurred())

		utxos := btcutxo.UTXOs{
			btcutxo.NewStandardUTXO(
				client.Network().Chain(),
				"4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1",
				20000,
				"",
				0,
				0,
			),
			btcutxo.NewStandardUTXO(
				client.Network().Chain(),
				"1b112299cc23d43935619437343d56e75fd278af50aaa79e0839bb4d08e0e9b4",
				80000,
				"",
				1,
				0,
			),
		}

		recipients := btcaddress.Recipients{
			{
				Address: recipient,
				Amount:  40000,
			},
		}

		tx, err := client.BuildUnsignedTx(utxos, recipients, address, 600)
		Expect(err).ToNot(HaveOccurred())

		return tx
	}
	for _, testCase := range testCases {
		testCase := testCase
		Context("when fetching UTXOs", func() {
			It("should return the UTXO for a transaction with unspent outputs", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())
				utxo, err := client.UTXO(testCase.UnspentTxHash, testCase.UnspentVout)
				Expect(err).NotTo(HaveOccurred())
				Expect(utxo.TxHash()).To(Equal(testCase.UnspentTxHash))
				Expect(utxo.Amount()).To(Equal(testCase.Amount))
				Expect(utxo.ScriptPubKey()).To(Equal(testCase.ScriptPubKey))
				Expect(utxo.Vout()).To(Equal(testCase.UnspentVout))
			})

			It("should return an error for an invalid UTXO index", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.UTXO(testCase.UnspentTxHash, 10)
				Expect(err).To(Equal(ErrUTXOSpent))
			})

			It("should return an error for a UTXO that has been spent", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.UTXO(testCase.SpentTxHash, testCase.SpentVout)
				Expect(err).To(Equal(ErrUTXOSpent))
			})

			It("should return an error for an invalid transaction hash", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.UTXO("abcdefg", 0)
				Expect(err).To(Equal(ErrInvalidTxHash))
			})

			It("should return an error for a non-existent transaction hash", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.UTXO(testCase.NonExistentTxHash, 0)
				Expect(err).To(Equal(ErrTxHashNotFound))
			})

			It("should return a non-zero number of UTXOs for a funded address that has been imported", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())
				address, err := loadTestAccounts(testCase.Network).Address(44, 1, 0, 0, 1)
				Expect(err).NotTo(HaveOccurred())

				utxos, err := client.UTXOsFromAddress(address)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(BeNumerically(">", 0))
			})

			It("should return zero UTXOs for a randomly generated address", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())
				address, err := testutil.RandomAddress(testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				utxos, err := client.UTXOsFromAddress(address)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(Equal(0))
			})

		})

		Context("when building a utx", func() {
			It("should return the expected serialized transaction", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())
				address, err := loadTestAccounts(client.Network()).Address(44, 1, 0, 0, 1)
				Expect(err).NotTo(HaveOccurred())

				// Build unsigned transaction.
				tx := tx(client, address)
				txBytes, err := tx.Serialize()
				Expect(err).NotTo(HaveOccurred())
				serializedTx := hex.EncodeToString(txBytes)
				Expect(testCase.ExpectedTx).To(Equal(serializedTx))
			})
		})
	}
})
