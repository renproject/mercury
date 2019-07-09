package hdutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/hex"

	. "github.com/renproject/mercury/hdutil"
	"github.com/renproject/mercury/types/btctypes"
)

var _ = Describe("hdutil key derivation", func() {
	const mnemonic = "movie middle bunker bullet bind asset plastic permit right alter dwarf fragile"
	const password = "password"
	const network = btctypes.Testnet

	Context("when deriving bip39 keys", func() {
		It("can derive the correct bip39 seed", func() {
			seed := DeriveSeed(mnemonic, password)
			hexSeed := hex.EncodeToString(seed)
			Expect(hexSeed).To(Equal("f19d6e93a3deef23a02024f88681599cfe9fb73eb22a9e86125a50827eeee1a354f9567776aa25b03e352b03a82f3cc5369faf2bd8cc4599f1c1cecebec18e84"))
		})

		It("can derive the correct bip39 root key", func() {
			key, err := DeriveExtendedPrivKey(mnemonic, password, network)
			Expect(err).NotTo(HaveOccurred())
			Expect(key.String()).To(Equal("tprv8ZgxMBicQKsPdK2FAhXoEqVmCq4MuTssfYGkLo5jRBv6m7Wcdt97YWSEx5LDMZKnJHipswU7AzCGT7Zc7oxitD2vaAvQQcPRThYBzdJPCZw"))
		})
	})

	Context("when deriving bip44 paths", func() {
		It("can correctly derive private key for: m/44'/1'/0'/0/0", func() {
			key, err := DeriveExtendedPrivKey(mnemonic, password, network)
			Expect(err).NotTo(HaveOccurred())
			privKey, err := DerivePrivKey(key, 44, 1, 0, 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(privKey.D.String()).To(Equal("cPzprCaAHhVxxk1hD5ua6nPtnjCcv4j5zEZ5kDKMJ2NB7gvfjX78"))
		})

		It("can correctly derive private key for: m/44'/1'/0'/0/1", func() {
			key, err := DeriveExtendedPrivKey(mnemonic, password, network)
			Expect(err).NotTo(HaveOccurred())
			privKey, err := DerivePrivKey(key, 44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(privKey.D.String()).To(Equal("cR32qhUyB2z7pnKhNUE7QEetUNPZD7GcSv5b7qAYhp8Rt8fH7UrW"))
		})

		It("can correctly derive private key for: m/44'/1'/0'/0/2", func() {
			key, err := DeriveExtendedPrivKey(mnemonic, password, network)
			Expect(err).NotTo(HaveOccurred())
			privKey, err := DerivePrivKey(key, 44, 1, 0, 0, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(privKey.D.String()).To(Equal("cTXTB37FgvZB9VimUTT9XMwegMotPRB4fHRZbw1XVLVYSWV7w4iB"))
		})
	})
})
