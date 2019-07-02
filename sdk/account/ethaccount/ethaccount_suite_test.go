package ethaccount_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/mercury/sdk/account/ethaccount"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/testutils"
)

func TestEthaccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethaccount Suite")
}

var Client ethclient.EthClient
var Account ethaccount.Account

var _ = BeforeSuite(func() {
	var err error
	Client, err = ethclient.NewCustomEthClient(fmt.Sprintf("http://localhost:%s", os.Getenv("GANACHE_PORT")))
	Expect(err).NotTo(HaveOccurred())
	key, ownerAddress, err := testutils.NewAccountFromHexPrivateKey(os.Getenv("LOCAL_ETH_TESTNET_PRIVATE_KEY"))
	Expect(err).NotTo(HaveOccurred())
	Account, err = ethaccount.NewAccountFromPrivateKey(Client, key)
	Expect(err).NotTo(HaveOccurred())
	ownerBal, err := Account.Balance(context.Background())
	Expect(err).NotTo(HaveOccurred())
	ownerBal2, err := Client.Balance(context.Background(), ownerAddress)
	Expect(ownerBal.Eq(ownerBal2)).Should(BeTrue())
})
