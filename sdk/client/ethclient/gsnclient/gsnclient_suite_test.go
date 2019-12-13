package gsnclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGsnclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gsnclient Suite")
}
