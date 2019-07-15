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
	"github.com/renproject/mercury/rpc/zecrpc"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutil"
	"github.com/renproject/mercury/testutil/btcaccount"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btcutxo"

	"github.com/sirupsen/logrus"
)

var _ = Describe("btc gateway", func() {
	logger := logrus.StandardLogger()

	// loadTestAccounts loads a HD Extended key for this tests. Some addresses of certain path has been set up for this
	// test. (i.e have known balance, utxos.)
	loadTestAccounts := func(network btctypes.Network) testutil.HdKey {
		switch network.Chain() {
		case types.Bitcoin:
			wallet, err := testutil.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE", network)
			Expect(err).NotTo(HaveOccurred())
			return wallet
		case types.ZCash:
			wallet, err := testutil.LoadHdWalletFromEnv("ZEC_TEST_MNEMONIC", "ZEC_TEST_PASSPHRASE", network)
			Expect(err).NotTo(HaveOccurred())
			return wallet
		default:
			panic(types.ErrUnknownChain)
		}
	}

	BeforeSuite(func() {
		btcStore := kv.NewJSON(kv.NewMemDB())
		btcCache := cache.New(btcStore, logger)

		zecStore := kv.NewJSON(kv.NewMemDB())
		zecCache := cache.New(zecStore, logger)

		btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
		btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
		btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
		btcTestnetNodeClient, err := btcrpc.NewNodeClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
		Expect(err).ToNot(HaveOccurred())

		zecTestnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
		zecTestnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
		zecTestnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")
		zecTestnetNodeClient, err := zecrpc.NewNodeClient(zecTestnetURL, zecTestnetUser, zecTestnetPassword)
		Expect(err).ToNot(HaveOccurred())

		btcTestnetAPI := api.NewBtcApi(btctypes.BtcTestnet, proxy.NewProxy(btcTestnetNodeClient), btcCache, logger)
		zecTestnetAPI := api.NewZecApi(btctypes.ZecTestnet, proxy.NewProxy(zecTestnetNodeClient), zecCache, logger)

		server := api.NewServer(logger, "5000", btcTestnetAPI, zecTestnetAPI)
		go server.Run()
	})

	networks := []btctypes.Network{btctypes.BtcTestnet, btctypes.ZecTestnet}
	Context("when generating gateways", func() {
		for _, network := range networks {
			network := network
			It("should be able to generate a btc gateway", func() {
				client, err := btcclient.New(logger, network)
				Expect(err).NotTo(HaveOccurred())
				key, err := loadTestAccounts(network).EcdsaKey(44, 1, 0, 0, 1)
				gateway := New(client, &key.PublicKey, []byte{})
				account, err := btcaccount.NewAccount(client, key)
				Expect(err).NotTo(HaveOccurred())

				// Transfer some funds to the gateway address
				amount := 20000 * btctypes.SAT

				// Fund mjSUANWKvokgHo6mxoHdq27aBgdCJ39uNA if the following transfer fails with not enough balance.
				txHash, err := account.Transfer(gateway.Address(), amount, types.Standard)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("funding gateway address=%v with txhash=%v\n", gateway.Address(), txHash)
				// Sleep for a small period of time in hopes that the transaction will go through
				time.Sleep(5 * time.Second)

				// Fetch the UTXOs for the transaction hash
				gatewayUTXO, err := gateway.UTXO(txHash, 1)
				Expect(err).NotTo(HaveOccurred())
				// fmt.Printf("utxo: %v", gatewayUTXO)
				gatewayUTXOs := btcutxo.UTXOs{gatewayUTXO}
				Expect(len(gatewayUTXOs)).To(BeNumerically(">", 0))
				txSize := gateway.EstimateTxSize(0, len(gatewayUTXOs), 1)
				gasAmount := client.SuggestGasPrice(context.Background(), types.Standard, txSize)
				fmt.Printf("gas amount=%v", gasAmount)
				recipients := btcaddress.Recipients{{
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
				serializedPK := btcaddress.SerializePublicKey(&key.PublicKey, client.Network())
				err = tx.InjectSignatures(sigs, serializedPK)

				Expect(err).NotTo(HaveOccurred())
				newTxHash, err := client.SubmitSignedTx(tx)
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("spending gateway funds with tx hash=%v\n", newTxHash)
			})
		}
	})
})
