package btcgateway_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBtcgateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Btcgateway Suite")
}
