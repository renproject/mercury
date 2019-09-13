package ethtypes_test

import (
	"crypto/rand"
	"encoding/hex"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth addresses", func() {

	Context("when calculating addresses", func() {
		It("should calcualte an address from ecdsa public key correctly", func() {
			key, err := testutil.RandomKey()
			Expect(err).Should(BeNil())
			Expect(func() { ethtypes.AddressFromPublicKey(&key.PublicKey) }).ShouldNot(Panic())
		})

		It("should calcualte an address from hex string correctly", func() {
			addrBytes := [20]byte{}
			rand.Read(addrBytes[:])
			Expect(func() { ethtypes.AddressFromHex(hex.EncodeToString(addrBytes[:])) }).ShouldNot(Panic())
		})
	})

})
