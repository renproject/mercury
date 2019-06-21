package ethclient_test

import (
	"context"
	"fmt"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/ethclient"

	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth client", func() {

	testAddress := func(network ethtypes.EthNetwork) ethtypes.EthAddr {
		var address ethtypes.EthAddr
		var err error
		switch network {
		case ethtypes.EthMainnet:
			address = ethtypes.HexStringToEthAddr("0xF02c1c8e6114b1Dbe8937a39260b5b0a374432bB")
		case ethtypes.EthKovan:
			address = ethtypes.HexStringToEthAddr("0xec58d8b8c3cc568e247fcf2dc96d221bac548dfc")
		default:
			Fail("unknown network")
		}
		Expect(err).NotTo(HaveOccurred())
		return address
	}

	for _, network := range []ethtypes.EthNetwork{ethtypes.EthMainnet, ethtypes.EthKovan} {
		network := network
		Context(fmt.Sprintf("when querying info of ethereum %s", network), func() {
			It("should return a non-zero balance", func() {
				client, err := NewEthClient(network)
				Expect(err).NotTo(HaveOccurred())
				address := testAddress(network)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				balance, err := client.Balance(ctx, address)
				Expect(err).NotTo(HaveOccurred())
				// fmt.Println(balance)
				Expect(balance.Gt(ethtypes.Wei(0))).Should(BeTrue())
			})

			It("should return a non-zero block number", func() {
				client, err := NewEthClient(network)
				Expect(err).NotTo(HaveOccurred())
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				blockNumber, err := client.BlockNumber(ctx)
				Expect(err).NotTo(HaveOccurred())
				// fmt.Println(blockNumber)
				Expect(blockNumber.Cmp(big.NewInt(0))).Should(Equal(1))
			})
		})
	}
})
