package btcrpc_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/rpc/btcrpc"

	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("btc rpc client", func() {

	// initNodeClient initialize a rpc client talking to the bitcoin node.
	initNodeClient := func(network btctypes.Network) Client {
		rpcHost := os.Getenv("BITCOIN_TESTNET_RPC_URL")
		rpcUser := os.Getenv("BITCOIN_TESTNET_RPC_USER")
		rpcPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")

		client, err := NewNodeClient(network, rpcHost, rpcUser, rpcPassword)
		Expect(err).NotTo(HaveOccurred())
		return client
	}

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func() testutils.HdKey {
		key, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE")
		Expect(err).NotTo(HaveOccurred())
		return key
	}

	Context("rpc client of btc full node", func() {
		It("should be able to get utxos", func() {
			network := btctypes.Testnet
			client := initNodeClient(network)
			address, err := loadTestAccounts().Address(network, 44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.GetUTXOs(address, 999999, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(utxos)).Should(Equal(3))
			Expect(utxos[0].Amount).Should(Equal(100000 * btctypes.Satoshi))
			Expect(utxos[0].TxHash).Should(Equal("5b37954895af2afc310ae1cbdd1233056072945fff449186a278a4f4fd42f7a7"))
			Expect(utxos[1].Amount).Should(Equal(1000000 * btctypes.Satoshi))
			Expect(utxos[1].TxHash).Should(Equal("801046d60d631b908fdcd8ab81ae1b7275bbb5a06aae57f1f1925de72483e1d4"))
			Expect(utxos[2].Amount).Should(Equal(1000000 * btctypes.Satoshi))
			Expect(utxos[2].TxHash).Should(Equal("375190ced26f4437bb5ef6766081f18cb730f6d0454612cb34d100db1a3626fb"))
		})

		It("should be able to get confirmation of a tx", func() {
			network := btctypes.Testnet
			client := initNodeClient(network)

			confirmations, err := client.Confirmations("5b37954895af2afc310ae1cbdd1233056072945fff449186a278a4f4fd42f7a7")
			Expect(err).NotTo(HaveOccurred())
			Expect(confirmations).Should(BeNumerically(">", 0))
		})
	})

	// Context ("testing ", func() {
	//
	// 	FIt("should", func() {
	// 		network := btctypes.Testnet
	//
	// 		client := initNodeClient(network)
	// 		address ,err  := btctypes.AddressFromBase58String("moJt8bi1t9sx4Pneb5v1QrfBzBkoyey6DH",network)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		utxos, err := client.GetUTXOs(address, 999999, 0)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		log.Print(utxos)
	// 	})
	// })
})
