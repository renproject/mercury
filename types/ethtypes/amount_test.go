package ethtypes_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth amounts", func() {

	Context("when comparing amounts", func() {

		It("should define WEI, GWEI, and ETHER correctly", func() {
			Expect(ethtypes.WEI.Lt(ethtypes.GWEI)).Should(BeTrue())
			Expect(ethtypes.WEI.Lte(ethtypes.GWEI)).Should(BeTrue())
			Expect(ethtypes.GWEI.Lt(ethtypes.ETHER)).Should(BeTrue())
			Expect(ethtypes.GWEI.Lte(ethtypes.ETHER)).Should(BeTrue())
			Expect(ethtypes.GWEI.Lte(ethtypes.GWEI)).Should(BeTrue())
			Expect(ethtypes.ETHER.Lte(ethtypes.ETHER)).Should(BeTrue())
			Expect(ethtypes.ETHER.Gte(ethtypes.ETHER)).Should(BeTrue())
			Expect(ethtypes.ETHER.Gte(ethtypes.GWEI)).Should(BeTrue())
			Expect(ethtypes.GWEI.Gte(ethtypes.WEI)).Should(BeTrue())
		})

		It("should correctly compare amounts", func() {
			zero := ethtypes.Wei(0)
			five := ethtypes.Wei(5)
			ten := ethtypes.Wei(10)
			Expect(ten.Gt(five)).Should(BeTrue())
			Expect(five.Gt(zero)).Should(BeTrue())
			Expect(zero.Eq(ethtypes.Wei(0))).Should(BeTrue())
			Expect(zero.Lt(five)).Should(BeTrue())
			Expect(five.Lt(ten)).Should(BeTrue())
		})
	})

	Context("when initiating amounts", func() {
		It("should initiate GWEI with the right decimals", func() {
			gwei1 := ethtypes.Wei(uint64(math.Pow(10, 9)))
			Expect(gwei1.Eq(ethtypes.GWEI)).Should(BeTrue())
			gwei2 := ethtypes.Gwei(1)
			Expect(gwei2.Eq(ethtypes.GWEI)).Should(BeTrue())
		})

		It("should initiate ETHER with the right decimals", func() {
			ether1 := ethtypes.Wei(uint64(math.Pow(10, 18)))
			Expect(ether1.Eq(ethtypes.ETHER)).Should(BeTrue())
			ether2 := ethtypes.Ether(1)
			Expect(ether2.Eq(ethtypes.ETHER)).Should(BeTrue())
		})
	})

	Context("when applying arithmetic operations on amounts", func() {
		It("can accurately add amounts", func() {
			Expect(ethtypes.Wei(3).Add(ethtypes.Wei(5)).Eq(ethtypes.Wei(8))).Should(BeTrue())
		})

		It("can accurately subtract amounts", func() {
			Expect(ethtypes.Wei(5).Sub(ethtypes.Wei(3)).Eq(ethtypes.Wei(2))).Should(BeTrue())
		})

		It("can accurately divide amounts", func() {
			Expect(ethtypes.Wei(8).Div(ethtypes.Wei(2)).Eq(ethtypes.Wei(4))).Should(BeTrue())
			Expect(ethtypes.Wei(9).Div(ethtypes.Wei(2)).Eq(ethtypes.Wei(4))).Should(BeTrue())
		})

		It("can accurately multiply amounts", func() {
			Expect(ethtypes.Wei(3).Mul(ethtypes.Wei(2)).Eq(ethtypes.Wei(6))).Should(BeTrue())
			Expect(ethtypes.Wei(5).Mul(ethtypes.Wei(2)).Eq(ethtypes.Wei(10))).Should(BeTrue())
		})
	})
})
