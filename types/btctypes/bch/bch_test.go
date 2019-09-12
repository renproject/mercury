package bch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/types/btctypes/bch"

	"github.com/btcsuite/btcd/chaincfg"
)

var _ = Describe("bch", func() {
	Context("when decoding testnet bitcoin cash addresses", func() {
		It("should decode a valid bitcoin cash address", func() {
			_, err := DecodeAddress("qrch9dvf9rc45p728n7d8p4r7n067jrdxgjyklgcg6", &chaincfg.TestNet3Params)
			Expect(err).Should(BeNil())
		})

		It("should decode a valid bitcoin cash address with a prefix", func() {
			_, err := DecodeAddress("bchtest:qrch9dvf9rc45p728n7d8p4r7n067jrdxgjyklgcg6", &chaincfg.TestNet3Params)
			Expect(err).Should(BeNil())
		})
	})
})
