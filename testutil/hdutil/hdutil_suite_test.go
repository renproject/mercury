package hdutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHdutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hdutil Suite")
}
