package btcclient_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"

	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("btc client", func() {

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func() testutils.HdKey{
		key, err:= testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE")
		Expect(err).NotTo(HaveOccurred())
		return key
	}

	// Fixme : currently not testing mainnet.
	for _, network := range []btctypes.Network{ /*types.Mainnet,*/ btctypes.Testnet} {
		network := network
		Context(fmt.Sprintf("when querying info of bitcoin %s", network), func() {
			It("should return the right balance", func() {
				client := NewBtcClient(network)
				address, err := loadTestAccounts().Address(network, 44, 1, 0, 0, 1)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				balance, err := client.Balance(ctx, address, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(balance).Should(Equal(100000 * btctypes.Satoshi))
			})

			It("should return the utxos of the given address", func() {
				client := NewBtcClient(network)
				address, err := loadTestAccounts().Address(network, 44, 1, 0, 0, 1)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				utxos, err := client.UTXOs(ctx, address, 999999, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(Equal(1))
				Expect(utxos[0].Amount).Should(Equal(100000 * btctypes.Satoshi))
				Expect(utxos[0].TxHash).Should(Equal("5b37954895af2afc310ae1cbdd1233056072945fff449186a278a4f4fd42f7a7"))
			})

			It("should return the confirmations of a tx", func() {
				client := NewBtcClient(network)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				hash :="5b37954895af2afc310ae1cbdd1233056072945fff449186a278a4f4fd42f7a7"
				confirmations, err := client.Confirmations(ctx, hash)
				Expect(err).NotTo(HaveOccurred())
				Expect(confirmations).Should(BeNumerically(">", 0))
			})
		})

		Context(fmt.Sprintf("when submitting stx to bitcoin %s", network), func() {
			PIt("should be able to send a stx", func() {
				client := NewBtcClient(network)
				key,err  := loadTestAccounts().EcdsaKey(44, 1, 0, 0, 2)
				Expect(err).NotTo(HaveOccurred())
				address, err := loadTestAccounts().Address(network, 44, 1, 0, 0, 2)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				utxos, err := client.UTXOs(ctx, address, 999999, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(BeNumerically(">=", 1))

				stx, err := testutils.GenerateSignedTx(network, key, address.String(), int64(utxos[0].Amount), utxos[0].TxHash)
				Expect(err).NotTo(HaveOccurred())

				log.Println("successfully sign the tx,", hex.EncodeToString(stx))
				Expect(client.SubmitSTX(ctx, stx)).Should(Succeed())
			})
		})
	}
})
