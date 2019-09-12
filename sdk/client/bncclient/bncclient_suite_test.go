package bncclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBncclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bncclient Suite")
}
