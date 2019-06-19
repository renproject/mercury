package ethclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEthclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethclient Suite")
}
