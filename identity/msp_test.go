package identity_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/s7techlab/hlf-sdk-go/identity/testdata/Org1MSP"
)

func TestIdentity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Identity suite")
}

var _ = Describe(`Cert`, func() {

	Context(`MSP`, func() {

		var (
			msp *identity.MSPConfig
			err error
		)

		It(`allow to load msp dir with all options`, func() {
			msp, err = identity.MSPFromPath(Org1MSP.ID, `testdata/Org1MSP`,
				identity.WithOUConfig(),
				identity.WithCertChain())

			Expect(err).NotTo(HaveOccurred())
			Expect(msp).NotTo(BeNil())

			Expect(msp.GetMSPIdentifier()).To(Equal(Org1MSP.ID))

			Expect(msp.Admins()).To(HaveLen(1))
			Expect(msp.Admins()[0].GetPEM()).To(Equal(Org1MSP.AdminCert))

			Expect(msp.Signer()).NotTo(BeNil())
			Expect(msp.Signer().GetPEM()).To(Equal(Org1MSP.SignCert))

			Expect(msp.CACerts()).To(HaveLen(1))
			Expect(identity.PEMEncode(msp.CACerts()[0].Raw)).To(Equal(Org1MSP.CACert))

			Expect(msp.IntermediateCerts()).To(HaveLen(0))

			ouConfig := msp.OUConfig()
			Expect(ouConfig).NotTo(BeNil())

			// configured in testdata/Org1MSP/config.yaml
			Expect(ouConfig.UnitIdentifiers).To(HaveLen(0))
			Expect(ouConfig.NodeOUs).NotTo(BeNil())

			Expect(ouConfig.NodeOUs.Enable).To(Equal(true))
			Expect(ouConfig.NodeOUs.ClientOuIdentifier.Certificate).To(Equal(Org1MSP.CACert))
			Expect(ouConfig.NodeOUs.PeerOuIdentifier.Certificate).To(Equal(Org1MSP.CACert))
			Expect(ouConfig.NodeOUs.AdminOuIdentifier.Certificate).To(Equal(Org1MSP.CACert))
			Expect(ouConfig.NodeOUs.OrdererOuIdentifier.Certificate).To(Equal(Org1MSP.CACert))

		})

		It(`serialize msp config`, func() {
			files, err := msp.Serialize()
			Expect(err).NotTo(HaveOccurred())

			//  2: admincert + cacert
			//  4: for each role in config.yaml
			//  + config.yaml = 7
			Expect(files).To(HaveLen(7))

			Expect(files[`admincerts/cert_0.pem`]).To(Equal(Org1MSP.AdminCert))
			Expect(files[`cacerts/cert_0.pem`]).To(Equal(Org1MSP.CACert))
		})

		It(`allow to create msp from FabricMSPConfig`, func() {
			msp, err = identity.MSPFromConfig(Org1MSP.FabricMSPConfig())
			Expect(err).NotTo(HaveOccurred())
			Expect(msp).NotTo(BeNil())
		})
	})
})
