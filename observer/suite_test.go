package observer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestObservers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Observers Suite")
}

var _ = BeforeSuite(func() {
	channelsBlocksPeerCommonTestBeforeSuit()
	channelsBlocksPeerParsedTestBeforeSuit()
})
