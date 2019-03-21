package mercury_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMercury(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mercury Suite")
}
