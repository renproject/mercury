package btcclient_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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
		mnemonicENV := fmt.Sprintf("%s_TEST_MNEMONIC", strings.ToUpper(network.Chain().String()))
		passphraseENV := fmt.Sprintf("%s_TEST_PASSPHRASE", strings.ToUpper(network.Chain().String()))
		wallet, err := testutil.LoadHdWalletFromEnv(mnemonicENV, passphraseENV, network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
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
			btctypes.BtcLocalnet,
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
			btctypes.ZecLocalnet,
			btctypes.NewOutPoint("41ec71582bc44fb9abc2c5d2009d1352e7df118def521b3b17c5bff86e5cfb46", 1),
			btctypes.NewOutPoint("41ec71582bc44fb9abc2c5d2009d1352e7df118def521b3b17c5bff86e5cfb46", 10),
			btctypes.NewOutPoint("e96953b5030f44686e71650d6cb71a83625059ad086f7fc7802775e22cef0f65", 0),
			btctypes.NewOutPoint("abcdefg", 0),
			btctypes.NewOutPoint("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1", 0),
			"76a9143735df7c4d831491ce9dc462e6f606f6faffb5ca88ac",
			100000000,
			"0400008085202f8902b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a9147927c1f59b258381973c2e9b88e1aa88170db1e888ac08e80000000000001976a91439758249eea8e82cc7554822abad0ba1c32a3d1588ac00000000809698000000000000000000000000",
		},
		{
			btctypes.BchLocalnet,
			btctypes.NewOutPoint("5d2986a6adbea7a17a6fbd60dfb15b51d2ddfaee41659dd0d4a8bc2601c81e73", 1),
			btctypes.NewOutPoint("5d2986a6adbea7a17a6fbd60dfb15b51d2ddfaee41659dd0d4a8bc2601c81e73", 10),
			btctypes.NewOutPoint("09431560c96a97f1504b7a90fc1c56978cca4abeec8a9fae60c66c4b74e2cfa6", 0),
			btctypes.NewOutPoint("abcdefg", 0),
			btctypes.NewOutPoint("4b9e0e80d4bb9380e97aaa05fa872df57e65d34373491653934d32cc992211b1", 0),
			"76a91406689f883f5ec936d5384d5f75beb16d0c5aeafa88ac",
			10000000,
			"0100000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a914bc6baeb5b0b5daa34c2318cc647a911dfe40f0b488ac08e80000000000001976a914a4e1dbf6f6c7404ee1d685e6128f449eb9ca263288ac00000000",
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
			),
			btctypes.NewUTXO(
				btctypes.NewOutPoint("1b112299cc23d43935619437343d56e75fd278af50aaa79e0839bb4d08e0e9b4", 1),
				80000,
				nil,
				0,
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

		Context("when getting confirmations of a txhash", func() {
			It("should return 0 if the txHash does not exist", func() {
				client := NewClient(logger, testCase.Network)
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				hash := [32]byte{}
				rand.Read(hash[:])
				conf, err := client.Confirmations(ctx, types.TxHash(fmt.Sprintf("%x", hash[:])))
				Expect(conf).Should(BeZero())
				Expect(err).ShouldNot(BeNil())
			})
		})

		Context(fmt.Sprintf("when fetching UTXOs on %s %s", testCase.Network.Chain(), testCase.Network), func() {
			It("should return the UTXO for a transaction with unspent outputs", func() {
				client := NewClient(logger, testCase.Network)
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
				client := NewClient(logger, testCase.Network)

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err := client.UTXO(ctx, testCase.InvalidOutPoint)
				_, ok := err.(ErrUTXOSpent)
				Expect(ok).To(BeTrue())
			})

			It("should return an error for a UTXO that has been spent", func() {
				client := NewClient(logger, testCase.Network)

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err := client.UTXO(ctx, testCase.SpentOutPoint)
				_, ok := err.(ErrUTXOSpent)
				Expect(ok).To(BeTrue())
			})

			It("should return an error for an invalid transaction hash", func() {
				client := NewClient(logger, testCase.Network)

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err := client.UTXO(ctx, testCase.InvalidTxHashOutPoint)
				_, ok := err.(ErrInvalidTxHash)
				Expect(ok).To(BeTrue())
			})

			It("should return an error for a non-existent transaction hash", func() {
				client := NewClient(logger, testCase.Network)

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err := client.UTXO(ctx, testCase.NonExistentTxHashOutPoint)
				_, ok := err.(ErrTxHashNotFound)
				Expect(ok).To(BeTrue())
			})
		})

		Context(fmt.Sprintf("when building a utx on %s %s", testCase.Network.Chain(), testCase.Network), func() {
			It("should return the expected serialized transaction", func() {
				client := NewClient(logger, testCase.Network)
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
