package zecrpc_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestZecrpc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zecrpc Suite")
}
