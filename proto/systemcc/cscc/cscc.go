package cscc

import _ "embed"

var (
	//go:embed cscc.swagger.json
	Swagger []byte

	ServiceDesc = _CSCCService_serviceDesc
)
