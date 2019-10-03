package btctypes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("btctypes", func() {
	Context("when decoding testnet zcash addresses", func() {
		It("should decode a valid zcash address", func() {
			address, err := DecodeAddress("tmXj1bXqHFU9toMhLnAwFad5JcehNNqGASy")
			Expect(err).Should(BeNil())
			Expect(address.String()).Should(Equal("tmXj1bXqHFU9toMhLnAwFad5JcehNNqGASy"))
		})
	})
})
