package btcclient_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/btcclient"

	"github.com/renproject/mercury/types"

	"github.com/sirupsen/logrus"
)

var _ = Describe("bitcoin tx gas", func() {
	Context("when getting tx gas of different speed tier", func() {
		It("should return the live data", func() {
			logger := logrus.StandardLogger()
			gas := NewBtcGasStation(logger, 5*time.Second)

			ctx := context.Background()
			fastGas, err := gas.GasRequired(ctx, types.Fast, 1)
			Expect(err).Should(BeNil())
			Expect(fastGas).Should(BeNumerically(">", 1))
			logger.Infof("fast gas = %v sat/byte", fastGas)

			standardGas, err := gas.GasRequired(ctx, types.Standard, 1)
			Expect(err).Should(BeNil())
			Expect(standardGas).Should(BeNumerically(">", 1))
			logger.Infof("standard gas = %v sat/byte", standardGas)

			slowGas, err := gas.GasRequired(ctx, types.Slow, 1)
			Expect(err).Should(BeNil())
			Expect(slowGas).Should(BeNumerically(">", 1))
			logger.Infof("slow gas = %v sat/byte", slowGas)

			Expect(fastGas >= standardGas).Should(BeTrue())
			Expect(standardGas >= slowGas).Should(BeTrue())
		})
	})
})
