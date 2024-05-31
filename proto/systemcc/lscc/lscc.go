package lscc

import _ "embed"

var (
	//go:embed lscc.swagger.json
	Swagger []byte

	ServiceDesc = _LSCCInvokeService_serviceDesc
)
