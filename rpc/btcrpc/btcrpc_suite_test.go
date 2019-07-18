package btcrpc_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBtcrpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Btcrpc Suite")
}
