package btctypes_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strings"
	"testing/quick"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/types/btctypes"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/types"
)

var _ = Describe("btc types", func() {
	validateAddress := func(address Address, network Network, segwit bool) bool {
		if network == BtcMainnet {
			if segwit {
				return strings.HasPrefix(address.EncodeAddress(), "bc1")
			}
			return strings.HasPrefix(address.EncodeAddress(), "1")
		}
		addr := address.EncodeAddress()
		if segwit {
			return strings.HasPrefix(address.EncodeAddress(), "tb1")
		}
		return strings.HasPrefix(addr, "m") || strings.HasPrefix(addr, "n")
	}

	for _, network := range []Network{BtcTestnet, BtcMainnet} {
		network := network
		for _, segwit := range []bool{true, false} {
			segwit := segwit
			Context(fmt.Sprintf("when generate new btc addresses of %v", network), func() {
				It("should be able to generate random address of given network", func() {
					test := func() bool {
						address, err := testutil.RandomAddress(network, segwit)
						Expect(err).NotTo(HaveOccurred())
						return validateAddress(address, network, segwit)
					}
					Expect(quick.Check(test, nil)).To(Succeed())
				})

				It("should be able to decode an address from string", func() {
					test := func() bool {
						randAddr, err := testutil.RandomAddress(network, segwit)
						Expect(err).NotTo(HaveOccurred())
						address, err := AddressFromBase58(randAddr.EncodeAddress(), network)
						Expect(err).NotTo(HaveOccurred())
						return validateAddress(address, network, segwit)
					}
					Expect(quick.Check(test, nil)).To(Succeed())
				})

				It("should be able to decode an address from public key", func() {
					test := func() bool {
						randKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
						Expect(err).NotTo(HaveOccurred())
						address, err := AddressFromPubKey(randKey.PublicKey, network, segwit)
						return validateAddress(address, network, segwit)
					}
					Expect(quick.Check(test, nil)).To(Succeed())
				})

			})
		}
	}

	Context("bitcoin amount ", func() {
		It("should be converted correctly", func() {
			Expect(1e8 * SAT).Should(Equal(BTC))
		})
	})

	Context("bitcoin networks", func() {
		It("should be able to parse network from a string", func() {
			testnet := "testnet"
			Expect(func() { NewNetwork(types.Bitcoin, testnet) }).ShouldNot(Panic())

			testnet3 := "testnet3"
			Expect(func() { NewNetwork(types.Bitcoin, testnet3) }).ShouldNot(Panic())

			mainnet := "mainnet"
			Expect(func() { NewNetwork(types.Bitcoin, mainnet) }).ShouldNot(Panic())

			unknownNetwork := func(network string) bool {
				Expect(func() { NewNetwork(types.Bitcoin, testnet) }).ShouldNot(Panic())
				return true
			}
			Expect(quick.Check(unknownNetwork, nil)).To(Succeed())
		})
	})
})
