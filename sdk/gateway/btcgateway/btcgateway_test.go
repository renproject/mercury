package btcgateway_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/gateway/btcgateway"

	"github.com/btcsuite/btcd/btcec"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/testutil/btcaccount"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc gateway", func() {
	logger := logrus.StandardLogger()

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutil.HdKey {
		mnemonicENV := fmt.Sprintf("%s_TEST_MNEMONIC", strings.ToUpper(network.Chain().String()))
		passphraseENV := fmt.Sprintf("%s_TEST_PASSPHRASE", strings.ToUpper(network.Chain().String()))
		wallet, err := testutil.LoadHdWalletFromEnv(mnemonicENV, passphraseENV, network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
	}

	testCases := []struct {
		Network        btctypes.Network
		GatewayAddress string
		SpenderAddress string
	}{
		{
			btctypes.BtcTestnet,
			"2NDzN2erre3RS7UY8ieY1LsiZ7AwNxGEnTU",
			"mzPbx28MWBrEZdXbjauVgf8UgnCm5hFdpf",
		},
		{
			btctypes.ZecTestnet,

			"t2V72Lwadq2hkH1rdcgckLBEitRUtcViUU1",
			"tmXymiPTMYvNjVbfG2xa9S3hsUiKvd1SvMS",
		},
		{
			btctypes.BchTestnet,

			"ppl7dtqf2zx0l0sln2r53mvlyc4see253ujcxsvvq8",
			"qqdwdwu3lpcpcusct56l3t5ptvw32tv7ns4ez829fa",
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		Context(fmt.Sprintf("locally validating %s gateways", testcase.Network.Chain()), func() {
			It(fmt.Sprintf("should be able to generate a %v gateway", testcase.Network), func() {
				client := btcclient.NewClient(logger, testcase.Network)
				key, err := loadTestAccounts(client.Network()).EcdsaKey(44, 1, 1, 0, 1)
				Expect(err).Should(BeNil())
				gateway := New(client, key.PublicKey, []byte{})
				Expect(gateway.Address().EncodeAddress()).Should(Equal(testcase.GatewayAddress))
			})

			It(fmt.Sprintf("should be able to generate a %v gateway", testcase.Network), func() {
				client := btcclient.NewClient(logger, testcase.Network)
				key, err := loadTestAccounts(client.Network()).EcdsaKey(44, 1, 1, 0, 1)
				Expect(err).Should(BeNil())
				gateway := New(client, key.PublicKey, []byte{})
				Expect(gateway.Spender().EncodeAddress()).Should(Equal(testcase.SpenderAddress))
			})

			It(fmt.Sprintf("should panic when trying to use an invalid pub key to generate %v gateway", testcase.Network), func() {
				client := btcclient.NewClient(logger, testcase.Network)
				Expect(func() { New(client, ecdsa.PublicKey{}, []byte{}) }).Should(Panic())
			})
		})
	}

	Context("when generating gateways", func() {
		networks := []btctypes.Network{btctypes.BtcLocalnet, btctypes.ZecLocalnet, btctypes.BchLocalnet}
		for _, network := range networks {
			network := network
			It(fmt.Sprintf("should be able to generate a %v gateway", network), func() {
				client := btcclient.NewClient(logger, network)
				key, err := loadTestAccounts(network).EcdsaKey(44, 1, 0, 0, 1)
				gateway := New(client, key.PublicKey, []byte{})
				account, err := btcaccount.NewAccount(client, key)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()

				fmt.Println(network, account.Address())
				// Transfer some funds to the gateway address
				amount := 20000 * btctypes.SAT

				txHash, err := account.Transfer(ctx, gateway.Address(), amount, types.Standard, false)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("funding gateway address=%v with txhash=%v\n", gateway.Address(), txHash)
				// Sleep for a small period of time in hopes that the transaction will go through
				time.Sleep(5 * time.Second)

				// Fetch the UTXOs for the transaction hash
				gatewayUTXO, err := gateway.UTXO(ctx, btctypes.NewOutPoint(txHash, 0))
				Expect(err).NotTo(HaveOccurred())
				// fmt.Printf("utxo: %v", gatewayUTXO)
				gatewayUTXOs := btctypes.UTXOs{gatewayUTXO}
				Expect(len(gatewayUTXOs)).To(BeNumerically(">", 0))
				txSize := gateway.EstimateTxSize(0, len(gatewayUTXOs), 1)
				gasAmount := client.SuggestGasPrice(ctx, types.Standard, txSize)
				fmt.Printf("gas amount=%v", gasAmount)
				recipients := btctypes.Recipients{{
					Address: account.Address(),
					Amount:  gatewayUTXOs.Sum() - gasAmount,
				}}
				tx, err := client.BuildUnsignedTx(gatewayUTXOs, recipients, account.Address(), gasAmount)
				Expect(err).NotTo(HaveOccurred())

				// Sign the transaction
				subScripts := tx.SignatureHashes()
				sigs := make([]*btcec.Signature, len(subScripts))
				for i, subScript := range subScripts {
					sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
					Expect(err).NotTo(HaveOccurred())
				}
				err = tx.InjectSignatures(sigs, key.PublicKey)

				Expect(err).NotTo(HaveOccurred())
				newTxHash, err := client.SubmitSignedTx(context.Background(), tx)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("spending gateway funds with tx hash=%v\n", newTxHash)
			})
		}
	})

	Context("when using SegWit gateways", func() {
		It(fmt.Sprintf("should be able to generate a %v gateway", btctypes.BtcLocalnet), func() {
			client := btcclient.NewClient(logger, btctypes.BtcLocalnet)
			key, err := loadTestAccounts(btctypes.BtcLocalnet).EcdsaKey(44, 1, 0, 0, 1)
			gateway := New(client, key.PublicKey, []byte{})
			account, err := btcaccount.NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())
			// Transfer some funds to the gateway address
			amount := 20000 * btctypes.SAT

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			segWitAddr := gateway.BaseScript().(*btctypes.BtcScript).SegWitaddress()
			// Fund mjSUANWKvokgHo6mxoHdq27aBgdCJ39uNA if the following transfer fails with not enough balance.
			txHash, err := account.Transfer(ctx, segWitAddr, amount, types.Standard, false)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("funding gateway address=%v with txhash=%v\n", segWitAddr, txHash)
			// Sleep for a small period of time in hopes that the transaction will go through
			time.Sleep(5 * time.Second)

			// Fetch the UTXOs for the transaction hash
			gatewayUTXO, err := gateway.UTXO(ctx, btctypes.NewOutPoint(txHash, 0))
			Expect(err).NotTo(HaveOccurred())
			// fmt.Printf("utxo: %v", gatewayUTXO)
			gatewayUTXOs := btctypes.UTXOs{gatewayUTXO}
			Expect(len(gatewayUTXOs)).To(BeNumerically(">", 0))
			txSize := gateway.EstimateTxSize(0, len(gatewayUTXOs), 1)
			gasAmount := client.SuggestGasPrice(ctx, types.Standard, txSize)
			fmt.Printf("gas amount=%v", gasAmount)
			recipients := btctypes.Recipients{{
				Address: account.Address(),
				Amount:  gatewayUTXOs.Sum() - gasAmount,
			}}
			tx, err := client.BuildUnsignedTx(gatewayUTXOs, recipients, account.Address(), gasAmount)
			Expect(err).NotTo(HaveOccurred())

			// Sign the transaction
			subScripts := tx.SignatureHashes()
			sigs := make([]*btcec.Signature, len(subScripts))
			for i, subScript := range subScripts {
				sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
				Expect(err).NotTo(HaveOccurred())
			}
			err = tx.InjectSignatures(sigs, key.PublicKey)

			Expect(err).NotTo(HaveOccurred())
			newTxHash, err := client.SubmitSignedTx(ctx, tx)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("spending gateway funds with tx hash=%v\n", newTxHash)
		})
	})
})
