package btctypes_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("UTXOs", func() {

	Context("when marshaling an outpoint", func() {
		It("should marshal and unmarshal into the same value", func() {
			outpoint := NewOutPoint("something", 10)

			data, err := json.Marshal(outpoint)
			Expect(err).ShouldNot(HaveOccurred())
			reconstructedOutPoint := NewOutPoint("", 0)
			Expect(json.Unmarshal(data, &reconstructedOutPoint)).ShouldNot(HaveOccurred())
			Expect(reconstructedOutPoint).To(Equal(outpoint))
		})
	})
})
