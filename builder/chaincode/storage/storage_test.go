package storage

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestChaincodeStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaincode storage test suite")
}
