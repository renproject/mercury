package ethaccount_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth client", func() {

	Context("can sign", func() {
		It("can create an unsigned transaction", func() {
			ctx := context.Background()
			amount := ethtypes.Ether(3)
			gasLimit := uint64(1000)
			gasPrice, err := Client.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, addr, err := testutils.NewAccount()
			Expect(err).NotTo(HaveOccurred())
			var data []byte
			_, err = Account.CreateUTX(ctx, addr, amount, gasLimit, gasPrice, data)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can sign an unsigned transaction", func() {
			ctx := context.Background()
			amount := ethtypes.Ether(3)
			gasLimit := uint64(1000)
			gasPrice, err := Client.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, addr, err := testutils.NewAccount()
			Expect(err).NotTo(HaveOccurred())
			var data []byte
			utx, err := Account.CreateUTX(ctx, addr, amount, gasLimit, gasPrice, data)
			Expect(err).NotTo(HaveOccurred())
			_, err = Account.SignUTX(ctx, utx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can transfer funds", func() {
			ctx := context.Background()
			amount := ethtypes.Ether(3)
			gasLimit := uint64(6721975)
			gasPrice, err := Client.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, addr, err := testutils.NewAccount()
			Expect(err).NotTo(HaveOccurred())
			bal, err := Client.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			Expect(bal.Eq(ethtypes.Wei(0))).Should(BeTrue())
			var data []byte
			utx, err := Account.CreateUTX(ctx, addr, amount, gasLimit, gasPrice, data)
			Expect(err).NotTo(HaveOccurred())
			stx, err := Account.SignUTX(ctx, utx)
			Expect(err).NotTo(HaveOccurred())
			err = Client.PublishSTX(ctx, stx)
			Expect(err).NotTo(HaveOccurred())
			// check new balance
			newBal, err := Client.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			Expect(newBal.Eq(amount)).Should(BeTrue())
		})

	})

	/*
		testAddress := func(network ethtypes.EthNetwork) ethtypes.Address {
			var address ethtypes.Address
			var err error
			switch network {
			case ethtypes.EthMainnet:
				address = ethtypes.HexStringToAddress("0xF02c1c8e6114b1Dbe8937a39260b5b0a374432bB")
			case ethtypes.EthKovan:
				address = ethtypes.HexStringToAddress("0xec58d8b8c3cc568e247fcf2dc96d221bac548dfc")
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
