package btcrpc_test

import (
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/rpc/btcrpc"

	"github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("btc rpc client", func() {

	initNodeClient := func() Client {
		// rpcHost := os.Getenv("BITCOIN_TESTNET_RPC_URL")
		// rpcUser := os.Getenv("BITCOIN_TESTNET_RPC_USER")
		// rpcPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")

		// Fixme : use env variables
		rpcHost := "34.225.229.216:18332"
		rpcUser := "Mx8x4bnxLzeTAObCLXcRBroVswFQDfGj"
		rpcPassword := "nYzAU8YwDeMh4Ynfbo5oJNHgapbs8i3j"

		client, err := NewNodeClient(btctypes.Testnet, rpcHost, rpcUser, rpcPassword)
		Expect(err).NotTo(HaveOccurred())
		return client
	}

	Context("rpc client of btc full node", func() {
		It("should be able to get utxos", func() {
			client := initNodeClient()
			address, err := btctypes.AddressFromBase58String("myCBiJUDAuyJLXUTGPekaw4PDCQqiKwcdy", btctypes.Testnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.GetUTXOs(address, 999999, 0)
			Expect(err).NotTo(HaveOccurred())

			log.Printf("get %v utxos", len(utxos))

			Expect(true).Should(BeTrue())

			for i := range utxos {
				log.Println(utxos[i])
			}

			client.Blockinfo()
		})

		It("should be able to get confirmation of a tx", func() {

		})
	})
})
