package ethaccount_test

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/sdk/account/ethaccount"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/ethtypes"
)

func TestEthaccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethaccount Suite")
}

var StartingBalance ethtypes.Amount
var Client ethclient.EthClient
var Account ethaccount.Account

var cmd *exec.Cmd

var _ = BeforeSuite(func() {
	var err error
	key, err := crypto.GenerateKey()
	Expect(err).NotTo(HaveOccurred())
	StartingBalance = ethtypes.Ether(100)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses
	cmd, err = testutils.StartGanacheServer(ctx, key, StartingBalance.ToBig())
	Expect(err).NotTo(HaveOccurred())
	// go cmd.Wait()
	Client, err = ethclient.NewCustomEthClient(fmt.Sprintf("http://localhost:%v", testutils.PORT))
	Expect(err).NotTo(HaveOccurred())
	Account = ethaccount.NewEthAccount(Client, key)
})

var _ = AfterSuite(func() {
	Expect(cmd).NotTo(BeNil())
	err := cmd.Process.Kill()
	Expect(err).NotTo(HaveOccurred())
})
