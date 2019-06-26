package btcaccount_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/account/btcaccount"

	"github.com/sirupsen/logrus"
)

var _ = Describe("bitcoin tx fee", func() {
	Context("when getting tx fee of different speed tier", func() {
		It("should return the live data", func() {
			logger:= logrus.StandardLogger()
			fee := NewBitcoinFee(logger, 5 * time.Second)

			fastFee := fee.RecommendedTxFee(Fast)
			Expect(fastFee).Should(BeNumerically(">", 1))
			logger.Infof("fast fee = %v sat/byte", fastFee)

			standardFee := fee.RecommendedTxFee(Standard)
			Expect(standardFee).Should(BeNumerically(">", 1))
			logger.Infof("standard fee = %v sat/byte", standardFee)

			slowFee := fee.RecommendedTxFee(Slow)
			Expect(slowFee).Should(BeNumerically(">", 1))
			logger.Infof("slow fee = %v sat/byte", slowFee)

			Expect(fastFee >= standardFee).Should(BeTrue())
			Expect(standardFee >= slowFee).Should(BeTrue())
		})
	})
})
