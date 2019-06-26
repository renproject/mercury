package ethaccount_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEthaccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethaccount Suite")
}
