package lifecycle

import _ "embed"

var (
	//go:embed lifecycle.swagger.json
	Swagger []byte

	ServiceDesc = _LifecycleService_serviceDesc
)
