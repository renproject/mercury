package erc20_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestErc20(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Erc20 Suite")
}
