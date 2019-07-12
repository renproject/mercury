package ethaccount_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/sdk/account/ethaccount"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/sdk/client/ethclient"
)

func TestEthaccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethaccount Suite")
}

var Client ethclient.Client
var EthAccount Account

var _ = BeforeSuite(func() {
	var err error
	Client, err = ethclient.NewCustomClient(fmt.Sprintf("http://localhost:%s", os.Getenv("GANACHE_PORT")))
	Expect(err).NotTo(HaveOccurred())
	privateKey, err := crypto.HexToECDSA(os.Getenv("LOCAL_ETH_TESTNET_PRIVATE_KEY"))
	Expect(err).ToNot(HaveOccurred())
	ownerAccount, err := NewAccountFromPrivateKey(Client, privateKey)
	Expect(err).NotTo(HaveOccurred())
	EthAccount, err = NewAccountFromPrivateKey(Client, ownerAccount.PrivateKey())
	Expect(err).NotTo(HaveOccurred())
	ownerBal, err := EthAccount.Balance(context.Background())
	Expect(err).NotTo(HaveOccurred())
	ownerBal2, err := Client.Balance(context.Background(), ownerAccount.Address())
	Expect(ownerBal.Eq(ownerBal2)).Should(BeTrue())
})
