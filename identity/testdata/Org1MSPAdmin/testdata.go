package Org1MSPAdmin

import _ "embed"

var (
	//go:embed cacerts/localhost-7054-ca-org1.pem
	CACert []byte

	//go:embed signcerts/cert.pem
	SignCert []byte
)
