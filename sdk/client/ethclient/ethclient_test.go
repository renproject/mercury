package ethclient_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/ethclient"

	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth client", func() {
	var client *EthClient

	Context("when fetching balances", func() {
		It("can fetch a zero balance address", func() {
			_, addr, err := testutils.NewAccount()
			Expect(err).NotTo(HaveOccurred())
			client, err = NewCustomEthClient("http://127.0.0.1:8545")
			Expect(err).NotTo(HaveOccurred())
			ctx := context.Background()
			balance, err := client.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			Expect(balance.Eq(ethtypes.Wei(0))).Should(BeTrue())
		})

		It("can suggest a gas price", func() {
			ctx := context.Background()
			_, err := client.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can create unsigned transactions", func() {
			ctx := context.Background()
			amount := ethtypes.Ether(3)
			nonce := uint64(1)
			gasLimit := uint64(1000)
			gasPrice, err := client.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, addr, err := testutils.NewAccount()
			Expect(err).NotTo(HaveOccurred())
			var data []byte
			_ = client.CreateUTX(nonce, addr, amount, gasLimit, gasPrice, data)
		})

	})

	/*
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
	*/
})
