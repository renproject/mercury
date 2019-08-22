package btcclient_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"

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
		passphraseENV := fmt.Sprintf("%s_TEST_MNEMONIC", strings.ToUpper(network.Chain().String()))
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
			"0200000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c000000000000434104ce7ba8e0401b992913859998a489bfc61c5a3fffe060c985acdcc3ee1450cd116faeedcb7b24db1fbe308d069402363d8dd1046f880c53fcd505c7bfb0674e0eac08e800000000000043410486cad80fd0c719c89fa386a5d00adca92d7e88a16c8532f42f027ebd9bcff5436d9486bdb2961df95aa77425529a3a866158a3858fd3efcef158350e0548b240ac00000000",
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
			"0400008085202f8902b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a9149bfffb66c4a705a6fc03384de4994d3567ed869c88ac08e80000000000001976a914f8d8dfd3e841a5edf07ed9b248d4d8a6cde48b9088ac00000000809698000000000000000000000000",
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
			"0100000002b1112299cc324d935316497343d3657ef52d87fa05aa7ae98093bbd4800e9e4b0000000000ffffffffb4e9e0084dbb39089ea7aa50af78d25fe7563d343794613539d423cc9922111b0100000000ffffffff02409c0000000000001976a91404fa72eb45b7f84e4c4c253c30925271fc32fdd688ac08e80000000000001976a914f7198fe5c6ba45fe4d45c9dc63378258a3a9405f88ac00000000",
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
		timeout := 10 * time.Second

		Context(fmt.Sprintf("when fetching UTXOs on %s %s", testCase.Network.Chain(), testCase.Network), func() {
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
				_, ok := err.(ErrUTXOSpent)
				Expect(ok).To(BeTrue())
			})

			It("should return an error for a UTXO that has been spent", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.SpentOutPoint)
				_, ok := err.(ErrUTXOSpent)
				Expect(ok).To(BeTrue())
			})

			It("should return an error for an invalid transaction hash", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.InvalidTxHashOutPoint)
				_, ok := err.(ErrInvalidTxHash)
				Expect(ok).To(BeTrue())
			})

			It("should return an error for a non-existent transaction hash", func() {
				client, err := New(logger, testCase.Network)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				_, err = client.UTXO(ctx, testCase.NonExistentTxHashOutPoint)
				_, ok := err.(ErrTxHashNotFound)
				Expect(ok).To(BeTrue())
			})
		})

		Context(fmt.Sprintf("when building a utx on %s %s", testCase.Network.Chain(), testCase.Network), func() {
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
