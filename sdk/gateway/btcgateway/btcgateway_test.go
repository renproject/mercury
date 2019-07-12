package btcgateway_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/btcd/btcec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/gateway/btcgateway"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/testutils/btcaccount"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc gateway", func() {
	logger := logrus.StandardLogger()

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutils.HdKey {
		wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", network)
		Expect(err).NotTo(HaveOccurred())
		return wallet
	}

	BeforeSuite(func() {
		store := kv.NewJSON(kv.NewMemDB())
		cache := cache.New(store, logger)

		btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
		btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
		btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
		btcTestnetNodeClient, err := btcrpc.NewNodeClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
		Expect(err).ToNot(HaveOccurred())

		btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
		btcTestnetAPI := api.NewBtcApi(btctypes.Testnet, btcTestnetProxy, cache, logger)
		server := api.NewServer(logger, "5000", btcTestnetAPI)
		go server.Run()
	})

	Context("when generating gateways", func() {
		It("should be able to generate a gateway", func() {
			client, err := btcclient.New(logger, btctypes.Localnet)
			Expect(err).NotTo(HaveOccurred())
			key, err := loadTestAccounts(btctypes.Localnet).EcdsaKey(44, 1, 0, 0, 1)
			gateway := New(client, &key.PublicKey, []byte{})
			account, err := btcaccount.NewAccount(client, key)
			Expect(err).NotTo(HaveOccurred())

			// Transfer some funds to the gateway address
			amount := 20000 * btctypes.SAT
			txHash, err := account.Transfer(gateway.Address(), amount, types.Standard)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("funding gateway address=%v with txhash=%v\n", gateway.Address(), txHash)
			// Sleep for a small period of time in hopes that the transaction will go through
			time.Sleep(5 * time.Second)

			// Fetch the UTXOs for the transaction hash
			gatewayUTXO, err := client.UTXO(txHash, 0)
			Expect(err).NotTo(HaveOccurred())
			// fmt.Printf("utxo: %v", gatewayUTXO)
			gatewayUTXOs := btctypes.UTXOs{gatewayUTXO}
			Expect(len(gatewayUTXOs)).To(BeNumerically(">", 0))
			txSize := gateway.EstimateTxSize(0, len(gatewayUTXOs), 2)
			gasStation := client.GasStation()
			gasAmount := gasStation.CalculateGasAmount(context.Background(), types.Standard, txSize)
			// fmt.Printf("gas amount=%v", gasAmount)
			utxos, err := account.UTXOs()
			Expect(err).NotTo(HaveOccurred())
			tx, err := gateway.BuildUnsignedTx(gatewayUTXOs, utxos, gasAmount)
			Expect(err).NotTo(HaveOccurred())

			// Sign the transaction
			subScripts := tx.SignatureHashes()
			sigs := make([]*btcec.Signature, len(subScripts))
			for i, subScript := range subScripts {
				sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
				Expect(err).NotTo(HaveOccurred())
			}
			serializedPK := btctypes.SerializePublicKey(&key.PublicKey, client.Network())
			err = tx.InjectSignatures(sigs, serializedPK)

			Expect(err).NotTo(HaveOccurred())
			newTxHash, err := client.SubmitSignedTx(tx)
			Expect(err).NotTo(HaveOccurred())
			fmt.Printf("spending gateway funds with tx hash=%v\n", newTxHash)
		})
	})
})
