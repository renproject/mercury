package erc20_test

import (
	"context"
	"fmt"
	"math/big"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/contract/erc20"

	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("ERC20 contract", func() {
	testcases := []struct {
		Network ethtypes.Network

		ContractAddress ethtypes.Address
		UserAddress     ethtypes.Address
	}{
		{
			ethtypes.Kovan,

			ethtypes.AddressFromHex("0x2cd647668494c1b15743ab283a0f980d90a87394"),
			ethtypes.AddressFromHex("0xaD34c12F6000B28a7C583EfC19D631735d1313c4"),
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		// TODO: Add tests for the other functions on the ERC20 contract
		Context("when interacting with an ERC20 contract", func() {
			It("should be able to call decimals on an ERC20 contract", func() {
				client, err := ethclient.New(logrus.StandardLogger(), ethtypes.Kovan)
				Expect(err).Should(BeNil())
				erc20, err := New(client, testcase.ContractAddress)
				Expect(err).Should(BeNil())
				decimals, err := erc20.Decimals(context.Background())
				Expect(err).Should(BeNil())
				Expect(decimals).Should(Equal(uint8(18)))
			})

			It("should be able to watch for events on an ERC20 contract", func() {
				client, err := ethclient.New(logrus.StandardLogger(), ethtypes.Kovan)
				Expect(err).Should(BeNil())
				erc20, err := New(client, testcase.ContractAddress)
				Expect(err).Should(BeNil())
				events := make(chan ethtypes.Event, 10)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				go erc20.Watch(ctx, events, nil)
				for event := range events {
					fmt.Println(event.Args["value"].(*big.Int))
				}
			})
		})
	}
})
