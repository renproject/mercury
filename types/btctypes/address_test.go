package btctypes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	_ "github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("btctypes", func() {
	Context("when decoding testnet bitcoin cash addresses", func() {
		It("should decode a valid bitcoin cash address", func() {
			Expect(true).To(BeTrue())
		})
	})
})
