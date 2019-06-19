package btcclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBtcClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BtcClient Suite")
}
