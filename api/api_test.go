package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/api"
)

var _ = Describe("APIs", func() {
	Context("when hashing requests", func() {
		It("should return the same hash for identical requests", func() {
			fstData := []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice","params":[]}`)
			fstHash, err := HashData(fstData)
			Expect(err).ToNot(HaveOccurred())

			sndData := []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice","params":[]}`)
			sndHash, err := HashData(sndData)
			Expect(err).ToNot(HaveOccurred())

			Expect(fstHash).To(Equal(sndHash))
		})
	})
})
