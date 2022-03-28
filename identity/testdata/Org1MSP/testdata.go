package Org1MSP

import (
	_ "embed"

	mspproto "github.com/hyperledger/fabric-protos-go/msp"
)

var (
	ID = `Org1MSP`

	//go:embed admincerts/cert.pem
	AdminCert []byte

	//go:embed cacerts/localhost-7054-ca-org1.pem
	CACert []byte

	//go:embed signcerts/cert.pem
	SignCert []byte

	//go:embed config.yaml
	ConfigYaml []byte
)

func FabricMSPConfig() *mspproto.FabricMSPConfig {
	return &mspproto.FabricMSPConfig{
		Name:                          ID,
		RootCerts:                     [][]byte{CACert},
		IntermediateCerts:             [][]byte{},
		Admins:                        [][]byte{AdminCert},
		RevocationList:                nil,
		SigningIdentity:               nil,
		OrganizationalUnitIdentifiers: nil,
		CryptoConfig:                  nil,
		TlsRootCerts:                  nil,
		TlsIntermediateCerts:          nil,
		// same as config yaml
		FabricNodeOus: &mspproto.FabricNodeOUs{
			Enable: true,
			ClientOuIdentifier: &mspproto.FabricOUIdentifier{
				Certificate:                  CACert,
				OrganizationalUnitIdentifier: "client",
			},
			PeerOuIdentifier: &mspproto.FabricOUIdentifier{
				Certificate:                  CACert,
				OrganizationalUnitIdentifier: "peer",
			},
			AdminOuIdentifier: &mspproto.FabricOUIdentifier{
				Certificate:                  CACert,
				OrganizationalUnitIdentifier: "admin",
			},
			OrdererOuIdentifier: &mspproto.FabricOUIdentifier{
				Certificate:                  CACert,
				OrganizationalUnitIdentifier: "orderer",
			},
		},
	}
}
