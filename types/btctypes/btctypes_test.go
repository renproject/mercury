package btctypes_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing/quick"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/types/btctypes"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/types"
)

var _ = Describe("btc types", func() {
	validateAddress := func(address Address, network Network) bool {
		return address.IsForNet(network.Params())
	}

	for _, network := range []Network{BtcTestnet, BtcMainnet} {
		network := network

		Context(fmt.Sprintf("when generating new %s addresses of %v", network.Chain(), network), func() {
			It("should be able to generate an address for the given network", func() {
				test := func() bool {
					address, err := testutil.RandomAddress(network)
					Expect(err).NotTo(HaveOccurred())
					return validateAddress(address, network)
				}
				Expect(quick.Check(test, nil)).To(Succeed())
			})

			It("should be able to generate a SegWit address for the given network", func() {
				test := func() bool {
					address, err := testutil.RandomSegWitAddress(network)
					Expect(err).NotTo(HaveOccurred())
					return validateAddress(address, network)
				}
				Expect(quick.Check(test, nil)).To(Succeed())
			})

			It("should be able to decode an address from string", func() {
				test := func() bool {
					randAddr, err := testutil.RandomAddress(network)
					Expect(err).NotTo(HaveOccurred())
					address, err := AddressFromBase58(randAddr.EncodeAddress(), network)
					Expect(err).NotTo(HaveOccurred())
					return validateAddress(address, network)
				}
				Expect(quick.Check(test, nil)).To(Succeed())
			})

			It("should be able to decode a SegWit address from string", func() {
				test := func() bool {
					randAddr, err := testutil.RandomSegWitAddress(network)
					Expect(err).NotTo(HaveOccurred())
					address, err := AddressFromBase58(randAddr.EncodeAddress(), network)
					Expect(err).NotTo(HaveOccurred())
					return validateAddress(address, network)
				}
				Expect(quick.Check(test, nil)).To(Succeed())
			})

			It("should be able to decode an address from public key", func() {
				test := func() bool {
					randKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
					Expect(err).NotTo(HaveOccurred())
					address, err := AddressFromPubKey(randKey.PublicKey, network)
					return validateAddress(address, network)
				}
				Expect(quick.Check(test, nil)).To(Succeed())
			})

			It("should be able to decode a SegWit address from public key", func() {
				test := func() bool {
					randKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
					Expect(err).NotTo(HaveOccurred())
					address, err := SegWitAddressFromPubKey(randKey.PublicKey, network)
					return validateAddress(address, network)
				}
				Expect(quick.Check(test, nil)).To(Succeed())
			})
		})

	}

	Context("bitcoin amount ", func() {
		It("should be converted correctly", func() {
			Expect(1e8 * SAT).Should(Equal(BTC))
		})
	})

	Context("bitcoin compat address", func() {
		It("should decode the address correctly", func() {
			addr, err := DecodeAddress("tmXj1bXqHFU9toMhLnAwFad5JcehNNqGASy")
			Expect(err).Should(BeNil())
			Expect(addr.String()).Should(Equal("tmXj1bXqHFU9toMhLnAwFad5JcehNNqGASy"))
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
