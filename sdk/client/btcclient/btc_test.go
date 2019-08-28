package btcclient_test

import (
	"context"
	"encoding/hex"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types"

	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/types/btctypes"
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

		UnspentOutPoint           btctypes.OutPoint
		InvalidOutPoint           btctypes.OutPoint
		SpentOutPoint             btctypes.OutPoint
		InvalidTxHashOutPoint     btctypes.OutPoint
		NonExistentTxHashOutPoint btctypes.OutPoint

		ScriptPubKey string
		Amount       btctypes.Amount

		ExpectedTx string
	}{
		{
			btctypes.BtcTestnet,
			btctypes.NewOutPoint("bd4bb310b0c6c4e5225bc60711931552e5227c94ef7569bfc7037f014d91030c", 0),
			btctypes.NewOutPoint("bd4bb310b0c6c4e5225bc60711931552e5227c94ef7569bfc7037f014d91030c", 10),
			btctypes.NewOutPoint("7e65d34373491653934d32cc992211b14b9e0e80d4bb9380e97aaa05fa872df5", 0),
			btctypes.NewOutPoint("abcdefg", 0),
			btctypes.NewOutPoint("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1", 0),
			"76a9142d2b683141de54613e7c6648afdb454fa3b4126d88ac",
			100000,
			"0200000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a914eb32aacf85fb8372fdd0e6f3cca4f9216e85f37288ac08e80000000000001976a91444458029b9de2280c67e6e0373a2ba946984960388ac00000000",
		},
		{
			btctypes.ZecTestnet,
			btctypes.NewOutPoint("41ec71582bc44fb9abc2c5d2009d1352e7df118def521b3b17c5bff86e5cfb46", 1),
			btctypes.NewOutPoint("41ec71582bc44fb9abc2c5d2009d1352e7df118def521b3b17c5bff86e5cfb46", 10),
			btctypes.NewOutPoint("e96953b5030f44686e71650d6cb71a83625059ad086f7fc7802775e22cef0f65", 0),
			btctypes.NewOutPoint("abcdefg", 0),
			btctypes.NewOutPoint("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1", 0),
			"76a9143735df7c4d831491ce9dc462e6f606f6faffb5ca88ac",
			100000000,
			"0400008085202f8902b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a9147927c1f59b258381973c2e9b88e1aa88170db1e888ac08e80000000000001976a91439758249eea8e82cc7554822abad0ba1c32a3d1588ac00000000809698000000000000000000000000",
		},
	}

	tx := func(client Client, address btctypes.Address) btctypes.BtcTx {
		recipient, err := loadTestAccounts(client.Network()).Address(44, 1, 0, 0, 2)
		Expect(err).NotTo(HaveOccurred())

		utxos := btctypes.UTXOs{
			btctypes.NewUTXO(
				btctypes.NewOutPoint("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1", 0),
				20000,
				nil,
				0,
				nil,
				nil,
			),
			btctypes.NewUTXO(
				btctypes.NewOutPoint("1b112299cc23d43935619437343d56e75fd278af50aaa79e0839bb4d08e0e9b4", 1),
				80000,
				nil,
				0,
				nil,
				nil,
			),
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
	for _, testCase := range testCases {
		testCase := testCase
		timeout := time.Second

		Context("when fetching UTXOs", func() {
			It("should return the UTXO for a transaction with unspent outputs", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				utxo, err := client.UTXO(ctx, testCase.UnspentOutPoint)
				Expect(err).NotTo(HaveOccurred())
				Expect(utxo.TxHash()).To(Equal(testCase.UnspentOutPoint.TxHash()))
				Expect(utxo.Amount()).To(Equal(testCase.Amount))
				Expect(hex.EncodeToString(utxo.ScriptPubKey())).To(Equal(testCase.ScriptPubKey))
				Expect(utxo.Vout()).To(Equal(testCase.UnspentOutPoint.Vout()))
			})

			It("should return an error for an invalid UTXO index", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.InvalidOutPoint)
				Expect(err).To(Equal(ErrUTXOSpent))
			})

			It("should return an error for a UTXO that has been spent", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.SpentOutPoint)
				Expect(err).To(Equal(ErrUTXOSpent))
			})

			It("should return an error for an invalid transaction hash", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.InvalidTxHashOutPoint)
				Expect(err).To(Equal(ErrInvalidTxHash))
			})

			It("should return an error for a non-existent transaction hash", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.NonExistentTxHashOutPoint)
				Expect(err).To(Equal(ErrTxHashNotFound))
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
