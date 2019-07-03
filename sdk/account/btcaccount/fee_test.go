package btcaccount_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/account/btcaccount"

	"github.com/sirupsen/logrus"
)

var _ = Describe("bitcoin tx fee", func() {
	Context("when getting tx fee of different speed tier", func() {
		It("should return the live data", func() {
			logger := logrus.StandardLogger()
			fee := NewBitcoinGas(logger, 5*time.Second)

			ctx := context.Background()
			fastGas := fee.GasRequired(ctx, Fast)
			Expect(fastGas).Should(BeNumerically(">", 1))
			logger.Infof("fast fee = %v sat/byte", fastGas)

			standardGas := fee.GasRequired(ctx, Standard)
			Expect(standardGas).Should(BeNumerically(">", 1))
			logger.Infof("standard fee = %v sat/byte", standardGas)

			slowGas := fee.GasRequired(ctx, Slow)
			Expect(slowGas).Should(BeNumerically(">", 1))
			logger.Infof("slow fee = %v sat/byte", slowGas)

			Expect(fastGas >= standardGas).Should(BeTrue())
			Expect(standardGas >= slowGas).Should(BeTrue())
		})
	})
})
