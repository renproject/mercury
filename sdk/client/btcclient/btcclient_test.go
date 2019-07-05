package btcclient_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"

	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("btc client", func() {

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutils.HdKey {
		wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
	}

	// Fixme : currently not testing mainnet.
	for _, network := range []btctypes.Network{ /*types.Mainnet,*/ btctypes.Testnet} {
		network := network
		Context(fmt.Sprintf("when fetching UTXOs on %s", network), func() {
			It("should return a non-zero number of UTXOs from the funded address", func() {
				client, err := NewBtcClient(network)
				Expect(err).NotTo(HaveOccurred())
				address, err := loadTestAccounts(network).Address(44, 1, 0, 0, 1)
				Expect(err).NotTo(HaveOccurred())

				utxos, err := client.UTXOs(address)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(BeNumerically(">", 0))
			})

			It("should return zero UTXOs from a randomly generated address", func() {
				client, err := NewBtcClient(network)
				Expect(err).NotTo(HaveOccurred())
				address, err := testutils.RandomAddress(network)
				Expect(err).NotTo(HaveOccurred())

				utxos, err := client.UTXOs(address)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(Equal(0))
			})
		})

		Context(fmt.Sprintf("when building a utx on %s", network), func() {
			PIt("should have the correct inputs", func() {
				// TODO: write the test
			})

			PIt("should have the correct outputs", func() {
				// TODO: write the test
			})
		})

		Context(fmt.Sprintf("when submitting stx to bitcoin %s", network), func() {
			PIt("should be able to submit a stx", func() {
				// TODO: write the test
			})
		})
	}
})
