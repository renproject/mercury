package ethclient_test

import (
	"context"
	"crypto/ecdsa"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/testutils"
	"github.com/renproject/mercury/types/ethtypes"
)

func TestEthclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethclient Suite")
}

var StartingBalance ethtypes.Amount
var Key *ecdsa.PrivateKey

var cmd *exec.Cmd

var _ = BeforeSuite(func() {
	var err error
	Key, err = crypto.GenerateKey()
	Expect(err).NotTo(HaveOccurred())
	StartingBalance = ethtypes.Ether(100)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses
	cmd, err = testutils.StartGanacheServer(ctx, Key, StartingBalance.ToBig())
	Expect(err).NotTo(HaveOccurred())
	// go cmd.Wait()
})

var _ = AfterSuite(func() {
	Expect(cmd).NotTo(BeNil())
	err := cmd.Process.Kill()
	Expect(err).NotTo(HaveOccurred())
})
