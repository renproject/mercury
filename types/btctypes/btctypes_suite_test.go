package btctypes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBtctypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Btctypes Suite")
}
