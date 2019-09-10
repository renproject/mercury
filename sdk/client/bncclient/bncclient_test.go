package bncclient_test

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/client/bncclient"
	"github.com/renproject/mercury/types/bnctypes"
)

var _ = Describe("bnc client", func() {
	Context("when communicating with the testnet", func() {
		It("should print the current time", func() {
			client := New(bnctypes.Testnet)
			client.PrintTime()
		})

		It("should get the correct balance of an address", func() {
			privKey, err := crypto.HexToECDSA(os.Getenv("BNB_PRIVATE_KEY"))
			Expect(err).Should(BeNil())
			address := bnctypes.AddressFromPubKey(privKey.PublicKey, bnctypes.Testnet)
			Expect(err).Should(BeNil())
			fmt.Println(address)
			client := New(bnctypes.Testnet)
			client.Balances(bnctypes.Address(address))
		})

		It("should transfer 0.001", func() {
			privKey, err := crypto.HexToECDSA(os.Getenv("BNB_PRIVATE_KEY"))
			Expect(err).Should(BeNil())
			address := bnctypes.AddressFromPubKey(privKey.PublicKey, bnctypes.Testnet)
			Expect(err).Should(BeNil())
			client := New(bnctypes.Testnet)
			tx, err := client.Send(address, bnctypes.Recipients{bnctypes.NewRecipent(address, bnctypes.NewBNBCoin(1000000))})
			Expect(err).Should(BeNil())
			btcPrivKey := (*btcec.PrivateKey)(privKey)
			sigs := []*btcec.Signature{}
			for _, hash := range tx.SignatureHashes() {
				sig, err := btcPrivKey.Sign(hash)
				if err != nil {
					Expect(err).Should(BeNil())
				}
				sigs = append(sigs, sig)
			}
			tx.InjectSignatures(sigs, privKey.PublicKey)
			Expect(client.SubmitTx(tx)).Should(BeNil())
		})
	})
})
