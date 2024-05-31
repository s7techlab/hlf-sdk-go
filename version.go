package hlf_sdk_go

import (
	"errors"
)

var (
	ErrUnknownFabricVersion = errors.New(`unknown fabric version`)
)

type FabricVersion string

const (
	FabricVersionUndefined FabricVersion = "undefined"
	FabricV1               FabricVersion = "1"
	FabricV2               FabricVersion = "2"
)

func FabricVersionIsV2(isV2 bool) FabricVersion {
	if isV2 {
		return FabricV2
	}

	return FabricV1
}
