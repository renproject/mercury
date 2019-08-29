package bncclient_test

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/bncclient"
	"github.com/renproject/mercury/types/bnctypes"
)

var _ = Describe("bnc client", func() {
	Context("when ", func() {
		It("should ", func() {
			client := New(bnctypes.Testnet)
			client.PrintTime()
			Expect(true).Should(BeTrue())
		})

		It("should ", func() {
			fmt.Println(os.Getenv("BNB_PRIVATE_KEY"))
			privKey, err := crypto.HexToECDSA(os.Getenv("BNB_PRIVATE_KEY"))
			Expect(err).Should(BeNil())
			address, err := bnctypes.AddressFromPubKey(privKey.PublicKey, bnctypes.Testnet)
			Expect(err).Should(BeNil())
			fmt.Println(address)
			client := New(bnctypes.Testnet)

			client.Balances(bnctypes.Address(address))
		})
	})
})
