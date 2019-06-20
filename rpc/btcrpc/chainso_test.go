package btcrpc_test

import (
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/rpc/btcrpc"

	"github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("chainso client", func() {

	initNodeClient := func() Client {
		return NewChainsoClient(btctypes.Testnet)
	}

	Context("rpc client of chainso API", func() {
		FIt("should be able to get utxos", func() {
			client := initNodeClient()
			address, err := btctypes.AddressFromBase58String("myCBiJUDAuyJLXUTGPekaw4PDCQqiKwcdy", btctypes.Testnet)
			Expect(err).NotTo(HaveOccurred())

			utxos, err := client.GetUTXOs(address, 999999, 0)
			Expect(err).NotTo(HaveOccurred())

			log.Printf("get %v utxos", len(utxos))
		})
	})
})
