package ethclient_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/ethclient"

	"github.com/renproject/mercury/types"

	"github.com/sirupsen/logrus"
)

var _ = Describe("ethereum tx gas", func() {
	Context("when getting tx gas of different speed tier", func() {
		It("should return the live data", func() {
			logger := logrus.StandardLogger()
			gas := NewethGasStation(logger, 5*time.Second)

			ctx := context.Background()
			fastGas := gas.GasRequired(ctx, types.Fast)
			Expect(fastGas).Should(BeNumerically(">", 1))
			logger.Infof("fast gas = %v sat/byte", fastGas)

			standardGas := gas.GasRequired(ctx, types.Standard)
			Expect(standardGas).Should(BeNumerically(">", 1))
			logger.Infof("standard gas = %v sat/byte", standardGas)

			slowGas := gas.GasRequired(ctx, types.Slow)
			Expect(slowGas).Should(BeNumerically(">", 1))
			logger.Infof("slow gas = %v sat/byte", slowGas)

			Expect(fastGas >= standardGas).Should(BeTrue())
			Expect(standardGas >= slowGas).Should(BeTrue())
		})
	})
})
