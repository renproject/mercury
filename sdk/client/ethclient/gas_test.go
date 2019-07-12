package ethclient_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"

	"github.com/renproject/mercury/types"

	"github.com/sirupsen/logrus"
)

var _ = Describe("ethereum tx gas", func() {
	Context("when getting tx gas of different speed tier", func() {
		It("should return the live data", func() {
			logger := logrus.StandardLogger()
			gas := NewEthGasStation(logger, 5*time.Second)

			ctx := context.Background()
			fastGas := gas.GasRequired(ctx, types.Fast)
			Expect(fastGas.Gt(ethtypes.Wei(1))).Should(BeTrue())
			logger.Infof("fast gas = %v", fastGas)

			standardGas := gas.GasRequired(ctx, types.Standard)
			Expect(standardGas.Gt(ethtypes.Wei(1))).Should(BeTrue())
			logger.Infof("standard gas = %v", standardGas)

			slowGas := gas.GasRequired(ctx, types.Slow)
			Expect(slowGas.Gt(ethtypes.Wei(1))).Should(BeTrue())
			logger.Infof("slow gas = %v", slowGas)

			Expect(fastGas.Gte(standardGas)).Should(BeTrue())
			Expect(standardGas.Gte(slowGas)).Should(BeTrue())
		})
	})
})
