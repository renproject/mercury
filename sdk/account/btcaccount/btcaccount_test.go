package btcaccount_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/account/btcaccount"

	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("btc account ", func() {
	Context("when transferring funds ", func() {
		PIt("should build the correct transaction and broadcast it", func() {
			client := btcclient.NewBtcClient(btctypes.Testnet)
			wallet, err := testutils.LoadHdWalletFromEnv("BTC_TEST_MNEMONIC", "BTC_TEST_PASSPHRASE")
			Expect(err).NotTo(HaveOccurred())
			key, err := wallet.EcdsaKey(44, 1, 0, 0, 1)
			Expect(err).NotTo(HaveOccurred())
			account := NewAccount(logrus.StandardLogger(), client, key)

			to, err := btctypes.AddressFromBase58("mhM9V7ENbJPpRnTGpVhNiHf631pzX2be74", btctypes.Testnet)
			Expect(err).NotTo(HaveOccurred())
			err = account.Transfer(context.Background(), to, 180000*btctypes.SAT, 0)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
