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
			gs := NewEthGasStation(logger, 5*time.Second)

			ctx := context.Background()
			fastGas, err := gs.GasRequired(ctx, types.Fast)
			Expect(err).NotTo(HaveOccurred())
			Expect(fastGas.Gt(ethtypes.Wei(1))).Should(BeTrue())
			logger.Infof("fast gas = %v", fastGas)

			standardGas, err := gs.GasRequired(ctx, types.Standard)
			Expect(err).NotTo(HaveOccurred())
			Expect(standardGas.Gt(ethtypes.Wei(1))).Should(BeTrue())
			logger.Infof("standard gas = %v", standardGas)

			slowGas, err := gs.GasRequired(ctx, types.Slow)
			Expect(err).NotTo(HaveOccurred())
			Expect(slowGas.Gt(ethtypes.Wei(1))).Should(BeTrue())
			logger.Infof("slow gas = %v", slowGas)

			Expect(fastGas.Gte(standardGas)).Should(BeTrue())
			Expect(standardGas.Gte(slowGas)).Should(BeTrue())
		})
	})
})
