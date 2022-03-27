package Org1MSP

import (
	_ "embed"
)

var (
	//go:embed admincerts/cert.pem
	AdminCert []byte

	//go:embed signcerts/cert.pem
	SignCert []byte

	//go:embed cacerts/localhost-7054-ca-org1.pem
	CACert []byte

	//go:embed config.yaml
	ConfigYaml []byte
)
