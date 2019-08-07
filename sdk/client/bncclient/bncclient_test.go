package bncclient_test

import (
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
	})
})
