package identity_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/s7techlab/hlf-sdk-go/identity"
)

func TestIdentity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Identity suite")
}

var _ = Describe(`Cert`, func() {

	Context(`MSP`, func() {

		It(`allow to load correct msp dir with all options`, func() {
			msp, err := identity.NewMSP(`Org1MSP`,
				identity.WithMSPPath(`testdata/Org1MSP`),
				identity.WithOUConfig(),
				identity.WithCertChain())

			Expect(err).NotTo(HaveOccurred())
			Expect(msp).NotTo(BeNil())

			Expect(msp.Signer()).NotTo(BeNil())
		})

	})
})
