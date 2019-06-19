package btcclient_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/mercury/types"
)

var _ = Describe("btc client", func() {

	testAddress := func(network types.BtcNetwork) types.BtcAddr {
		var address types.BtcAddr
		var err error
		switch network {
		case types.BtcMainnet:
			address, err = types.AddressFromBase58String("1MVC7MErbaqzgvXt647r7R9vy284HUJF5c", network)
		case types.BtcTestnet:
			address, err = types.AddressFromBase58String("mmmj7f5M1DK7Foq7oHejQYvmFCHdiRPk91", network)
		default:
			Fail("unknown network")
		}
		Expect(err).NotTo(HaveOccurred())
		return address
	}

	for _, network := range []types.BtcNetwork{ /*types.BtcMainnet,*/ types.BtcTestnet} {
		network := network
		Context(fmt.Sprintf("when querying info of bitcoin %s", network), func() {
			It("should return the right balance", func() {
				client := NewBtcClient(network)
				address := testAddress(network)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				balance, err := client.Balance(ctx, address, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(balance).Should(BeZero())
			})

			It("should return the utxos of the given address", func() {
				client := NewBtcClient(network)
				address := testAddress(network)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				utxos, err := client.UTXOs(ctx, address, 999999, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(utxos)).Should(BeZero())
			})

			It("should return the confirmations of a tx", func() {
				client := NewBtcClient(network)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				hash := types.TxHash("4b1f166b72d7838174c63aec75c27066fd1d9963982e22377d44ae485501c937")

				confirmations, err := client.Confirmations(ctx, hash)
				Expect(err).NotTo(HaveOccurred())
				Expect(confirmations).Should(BeNumerically(">", 0))
			})
		})

		Context(fmt.Sprintf("when submitting stx to bitcoin %s", network), func() {
			It("should be able to send a stx", func() {
				client := NewBtcClient(network)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				stx := []byte{}
				Expect(client.SubmitSTX(ctx, stx)).Should(Succeed())
			})
		})
	}
})
