package ethaccount_test

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/renproject/mercury/sdk/account/ethaccount"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("eth account", func() {

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
			gasLimit := uint64(30000)
			gasPrice, err := Client.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, addr, err := testutils.NewAccount()
			Expect(err).NotTo(HaveOccurred())
			bal, err := Client.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			Expect(bal.Eq(ethtypes.Wei(0))).Should(BeTrue())
			var data []byte
			utx, err := Account.CreateUTX(ctx, addr, amount, gasLimit, gasPrice, data)
			fmt.Println((*coretypes.Transaction)(utx).Hash())
			Expect(err).NotTo(HaveOccurred())
			stx, err := Account.SignUTX(ctx, utx)
			fmt.Println((*coretypes.Transaction)(stx).Hash())
			Expect(err).NotTo(HaveOccurred())
			err = Client.PublishSTX(ctx, stx)
			Expect(err).NotTo(HaveOccurred())
			// check new balance
			newBal, err := Client.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			Expect(newBal.Eq(amount)).Should(BeTrue())
		})

		It("can check kovan funds", func() {
			kovanClient, err := ethclient.NewCustomEthClient("http://139.59.221.34/eth-kovan")
			Expect(err).NotTo(HaveOccurred())
			mnemonic := os.Getenv("MNEMONIC_KOVAN")
			path := "m/44'/60'/0'/0/0"
			acc, err := ethaccount.NewAccountFromMnemonic(kovanClient, mnemonic, path)
			Expect(err).NotTo(HaveOccurred())
			ctx := context.Background()
			bal, err := kovanClient.Balance(ctx, acc.Address())
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("balance of %v: %v", acc.Address().Hex(), bal)
		})

		It("can send kovan funds", func() {
			kovanClient, err := ethclient.NewCustomEthClient("http://139.59.221.34/eth-kovan")
			Expect(err).NotTo(HaveOccurred())
			mnemonic := os.Getenv("MNEMONIC_KOVAN")
			path := "m/44'/60'/0'/0/0"
			acc, err := ethaccount.NewAccountFromMnemonic(kovanClient, mnemonic, path)
			Expect(err).NotTo(HaveOccurred())
			ctx := context.Background()
			amount := ethtypes.Ether(1)
			gasLimit := uint64(30000)
			gasPrice, err := kovanClient.SuggestGasPrice(ctx)
			Expect(err).NotTo(HaveOccurred())
			addr := ethtypes.HexStringToAddress("0xdF9dEfE40a4E3B2CfF85b51CfcBf87876C7Af902")
			bal, err := kovanClient.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("original balance: %v", bal)
			// Expect(bal.Eq(ethtypes.Wei(0))).Should(BeTrue())
			var data []byte
			utx, err := acc.CreateUTX(ctx, addr, amount, gasLimit, gasPrice, data)
			fmt.Println((*coretypes.Transaction)(utx).Hash())
			Expect(err).NotTo(HaveOccurred())
			stx, err := acc.SignUTX(ctx, utx)
			fmt.Println((*coretypes.Transaction)(stx).Hash())
			Expect(err).NotTo(HaveOccurred())
			err = kovanClient.PublishSTX(ctx, stx)
			Expect(err).NotTo(HaveOccurred())
			// check new balance
			newBal, err := kovanClient.Balance(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("new balance: %v", newBal)
			fmt.Printf("total balance: %v", bal.Add(amount))
			// Expect(newBal.Eq(bal.Add(amount))).Should(BeTrue())
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
