package ethclient_test

import (
	"context"
	"fmt"
	"math/big"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types"
	"github.com/sirupsen/logrus"

	"github.com/renproject/mercury/sdk/account/ethaccount"
	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth client", func() {
	var client Client
	var err error
	logger := logrus.StandardLogger()

	BeforeSuite(func() {
		client, err = NewCustomClient(logger, fmt.Sprintf("http://127.0.0.1:%v", os.Getenv("GANACHE_PORT")))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when fetching confirmations", func() {
		It("can fetch the confirmations of a Kovan transaction", func() {
			kovanClient, err := New(logger, ethtypes.Kovan)
			Expect(err).NotTo(HaveOccurred())
			hash := ethtypes.NewTxHashFromHex("0x288a0fe0cb305195bac6fefa6b16df576f0180c229fe5b4a453d57b0dcb42673")
			ctx := context.Background()
			confs, err := kovanClient.Confirmations(ctx, hash)
			Expect(err).NotTo(HaveOccurred())
			Expect(confs.Cmp(big.NewInt(0))).Should(BeEquivalentTo(1))
		})
	})

	Context("when fetching balances", func() {
		It("can fetch a zero balance address", func() {
			account, err := ethaccount.RandomAccount(client)
			Expect(err).NotTo(HaveOccurred())
			ctx := context.Background()
			balance, err := client.Balance(ctx, account.Address())
			Expect(err).NotTo(HaveOccurred())
			Expect(balance.Eq(ethtypes.Wei(0))).Should(BeTrue())
		})

		It("can check the gas limit", func() {
			ctx := context.Background()
			gl, err := client.GasLimit(ctx)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("gas limit: %v", gl)
		})

		It("can create unsigned transactions", func() {
			ctx := context.Background()
			amount := ethtypes.Ether(3)
			nonce := uint64(1)
			gasLimit := uint64(1000)
			gasPrice := client.SuggestGasPrice(ctx, types.Standard)
			account, err := ethaccount.RandomAccount(client)
			Expect(err).NotTo(HaveOccurred())
			var data []byte
			_, err = client.BuildUnsignedTx(ctx, nonce, account.Address(), amount, gasLimit, gasPrice, data)
			Expect(err).NotTo(HaveOccurred())
		})

	})

	/*
		testAddress := func(network ethtypes.Network) ethtypes.Address {
			var address ethtypes.Address
			var err error
			switch network {
			case ethtypes.Mainnet:
				address = ethtypes.HexStringToAddress("0xF02c1c8e6114b1Dbe8937a39260b5b0a374432bB")
			case ethtypes.Kovan:
				address = ethtypes.HexStringToAddress("0xec58d8b8c3cc568e247fcf2dc96d221bac548dfc")
			default:
				Fail("unknown network")
			}
			Expect(err).NotTo(HaveOccurred())
			return address
		}

		for _, network := range []ethtypes.Network{ethtypes.Mainnet, ethtypes.Kovan} {
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
