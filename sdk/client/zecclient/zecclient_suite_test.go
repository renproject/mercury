package zecclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestZecclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zecclient Suite")
}
