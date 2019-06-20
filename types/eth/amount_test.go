package eth_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/renproject/mercury/types/eth"
)

var _ = Describe("eth amounts", func() {

	Context("when comparing amounts", func() {
		It("should correctly compare amounts", func() {
			zero := eth.Wei(0)
			five := eth.Wei(5)
			ten := eth.Wei(10)
			Expect(ten.Gt(five)).Should(BeTrue())
			Expect(five.Gt(zero)).Should(BeTrue())
			Expect(zero.Eq(eth.Wei(0))).Should(BeTrue())
			Expect(zero.Lt(five)).Should(BeTrue())
			Expect(five.Lt(ten)).Should(BeTrue())
		})
	})

	Context("when initiating amounts", func() {
		It("should initiate GWEI with the right decimals", func() {
			gwei1 := eth.Wei(uint64(math.Pow(10, 9)))
			Expect(gwei1.Eq(eth.GWEI)).Should(BeTrue())
			gwei2 := eth.Gwei(1)
			Expect(gwei2.Eq(eth.GWEI)).Should(BeTrue())
		})

		It("should initiate ETHER with the right decimals", func() {
			ether1 := eth.Wei(uint64(math.Pow(10, 18)))
			Expect(ether1.Eq(eth.ETHER)).Should(BeTrue())
			ether2 := eth.Ether(1)
			Expect(ether2.Eq(eth.ETHER)).Should(BeTrue())
		})
	})

	Context("when adding amounts", func() {
		It("can accurately add amounts", func() {
			Expect(eth.Wei(3).Add(eth.Wei(5)).Eq(eth.Wei(8))).Should(BeTrue())
		})
	})
})
