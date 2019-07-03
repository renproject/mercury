package ethrpc_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEthrpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethrpc Suite")
}
