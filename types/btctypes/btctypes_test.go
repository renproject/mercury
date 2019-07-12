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
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
)

var _ = Describe("btc types ", func() {
	for _, network := range []Network{Testnet, Mainnet} {
		network := network
		Context(fmt.Sprintf("when generate new btc addresses of %v", network), func() {
			It("should be able to generate random address of given network", func() {
				randAddr := func() bool {
					address, err := testutils.RandomBTCAddress(network)
					Expect(err).NotTo(HaveOccurred())
					if network == Mainnet {
						return strings.HasPrefix(address.EncodeAddress(), "1")
					} else {
						addr := address.EncodeAddress()
						return strings.HasPrefix(addr, "m") || strings.HasPrefix(addr, "n")
					}
				}

				Expect(quick.Check(randAddr, nil)).To(Succeed())
			})

			It("should be able to decode an address from string", func() {
				randAddr, err := testutils.RandomBTCAddress(network)
				Expect(err).NotTo(HaveOccurred())
				address, err := btcaddress.BtcAddressFromBase58(randAddr.EncodeAddress(), network)
				Expect(err).NotTo(HaveOccurred())
				Expect(address.EncodeAddress()).Should(Equal(randAddr.EncodeAddress()))
			})

			It("should be able to decode an address from public key", func() {
				test := func() bool {
					randKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
					Expect(err).NotTo(HaveOccurred())
					address, err := btcaddress.BtcAddressFromPubKey(&randKey.PublicKey, network)
					if network == Mainnet {
						return strings.HasPrefix(address.EncodeAddress(), "1")
					} else {
						addr := address.EncodeAddress()
						return strings.HasPrefix(addr, "m") || strings.HasPrefix(addr, "n")
					}
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

	Context("bitcoin networks", func() {
		It("should be able to parse network from a string", func() {
			testnet := "testnet"
			Expect(func() { NewNetwork(testnet) }).ShouldNot(Panic())

			testnet3 := "testnet3"
			Expect(func() { NewNetwork(testnet3) }).ShouldNot(Panic())

			mainnet := "mainnet"
			Expect(func() { NewNetwork(mainnet) }).ShouldNot(Panic())

			unknownNetwork := func(network string) bool {
				Expect(func() { NewNetwork(testnet) }).ShouldNot(Panic())
				return true
			}
			Expect(quick.Check(unknownNetwork, nil)).To(Succeed())
		})
	})
})
