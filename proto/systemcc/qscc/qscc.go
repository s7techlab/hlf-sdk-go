package qscc

import _ "embed"

var (
	//go:embed qscc.swagger.json
	Swagger []byte

	ServiceDesc = _QSCCService_serviceDesc
)
