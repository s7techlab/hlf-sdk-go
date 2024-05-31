package wallet

import _ "embed"

var (
	//go:embed wallet.swagger.json
	Swagger []byte

	ServiceDesc = _WalletService_serviceDesc
)
