package bch_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bch Suite")
}
