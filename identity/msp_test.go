package identity_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/atomyze-ru/hlf-sdk-go/identity"
	"github.com/atomyze-ru/hlf-sdk-go/identity/testdata/Org1MSPAdmin"
	"github.com/atomyze-ru/hlf-sdk-go/identity/testdata/Org1MSPPeer"
)

func TestIdentity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Identity suite")
}

var _ = Describe(`Cert`, func() {

	Context(`MSP`, func() {

		Context(`Peer from path`, func() {

			var (
				msp *identity.MSPConfig
				err error
			)

			It(`allow to load peer msp from dir`, func() {
				msp, err = identity.MSPFromPath(Org1MSPPeer.ID, `testdata/Org1MSPPeer`)

				Expect(err).NotTo(HaveOccurred())
				Expect(msp).NotTo(BeNil())

				Expect(msp.GetMSPIdentifier()).To(Equal(Org1MSPPeer.ID))

				Expect(msp.Admins()).To(HaveLen(0))

				Expect(msp.Signer()).NotTo(BeNil())
				Expect(msp.Signer().GetPEM()).To(Equal(Org1MSPPeer.SignCert))

				mspConfig := msp.MSPConfig()
				Expect(mspConfig).NotTo(BeNil())

				Expect(mspConfig.RootCerts).To(HaveLen(1))
				Expect(mspConfig.RootCerts[0]).To(Equal(Org1MSPPeer.CACert))

				Expect(mspConfig.IntermediateCerts).To(HaveLen(0))

				// configured in testdata/Org1MSP/config.yaml
				Expect(mspConfig.OrganizationalUnitIdentifiers).To(HaveLen(0))
				Expect(mspConfig.FabricNodeOus).NotTo(BeNil())

				Expect(mspConfig.FabricNodeOus.Enable).To(Equal(true))
				Expect(mspConfig.FabricNodeOus.ClientOuIdentifier.Certificate).To(Equal(Org1MSPPeer.CACert))
				Expect(mspConfig.FabricNodeOus.PeerOuIdentifier.Certificate).To(Equal(Org1MSPPeer.CACert))
				Expect(mspConfig.FabricNodeOus.AdminOuIdentifier.Certificate).To(Equal(Org1MSPPeer.CACert))
				Expect(mspConfig.FabricNodeOus.OrdererOuIdentifier.Certificate).To(Equal(Org1MSPPeer.CACert))

			})

			It(`serialize msp config`, func() {
				files, err := msp.Serialize()
				Expect(err).NotTo(HaveOccurred())

				//  1: cacert
				//  4: for each role in config.yaml
				//  + config.yaml = 7
				Expect(files).To(HaveLen(6))
				Expect(files[`cacerts/cert_0.pem`]).To(Equal(Org1MSPPeer.CACert))

				Expect(files[`ou/admin.pem`]).To(Equal(Org1MSPPeer.CACert))
				Expect(files[`ou/peer.pem`]).To(Equal(Org1MSPPeer.CACert))
				Expect(files[`ou/client.pem`]).To(Equal(Org1MSPPeer.CACert))
				Expect(files[`ou/orderer.pem`]).To(Equal(Org1MSPPeer.CACert))

				Expect(files).To(HaveKey(`config.yaml`))
			})
		})

		Context(`Peer from FabricMSPCofig`, func() {

			It(`allow to create msp from FabricMSPConfig`, func() {
				msp, err := identity.MSPFromConfig(Org1MSPPeer.FabricMSPConfig())
				Expect(err).NotTo(HaveOccurred())
				Expect(msp).NotTo(BeNil())
			})

		})

		Context(`Peer + admin from path`, func() {
			It(`allow to load peer msp and admin msp from separate dirs`, func() {
				msp, err := identity.MSPFromPath(Org1MSPPeer.ID, `testdata/Org1MSPPeer`,
					identity.WithAdminMSPPath(`testdata/Org1MSPAdmin`))

				Expect(err).NotTo(HaveOccurred())
				Expect(msp).NotTo(BeNil())

				Expect(msp.Admins()).To(HaveLen(1))
				Expect(msp.Admins()[0].GetPEM()).To(Equal(Org1MSPAdmin.SignCert))

				Expect(msp.Signer()).NotTo(BeNil())
				Expect(msp.Signer().GetPEM()).To(Equal(Org1MSPPeer.SignCert))
			})

			It(`allow to load peer msp and admin msp from one dir (with admincerts subdir)`, func() {
				msp, err := identity.MSPFromPath(Org1MSPPeer.ID, `testdata/Org1MSPPeerAndAdmin`)

				Expect(err).NotTo(HaveOccurred())
				Expect(msp).NotTo(BeNil())

				Expect(msp.Admins()).To(HaveLen(1))
				Expect(msp.Admins()[0].GetPEM()).To(Equal(Org1MSPAdmin.SignCert))

				Expect(msp.Signer()).NotTo(BeNil())
				Expect(msp.Signer().GetPEM()).To(Equal(Org1MSPPeer.SignCert))
			})

		})

	})
})
