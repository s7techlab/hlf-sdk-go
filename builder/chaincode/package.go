package chaincode

import (
	"time"
)

type FabricVersion string

var (
	FabricV1          FabricVersion = "fabric-v1"
	FabricV2          FabricVersion = "fabric-v2"
	FabricV2Lifecycle FabricVersion = "fabric-v2-lifecycle"
)

type PackageID struct {
	// Name of chaincode.
	Name string `yaml:"name"`
	// Version of chaincode, default empty.
	Version string `yaml:"version"`
	// FabricVersion indicates which packager variant is used.
	FabricVersion FabricVersion `yaml:"fabric_version"`
}

type Package struct {
	PackageID `yaml:",inline"`
	// Repository is path to git sources, ex: http://:token@github.com/hyperledger-labs/cckit
	Repository string `yaml:"repository"`
	// ChaincodePath is path to chaincode, ex: github.com/hyperledger-labs/cckit/examples/erc2_utxo
	ChaincodePath string `yaml:"chaincode_path"`
	// BinaryPath is path to chaincode binary, ex: chaincode/bin
	BinaryPath string `yaml:"binary_path"`

	Source []byte `yaml:"-"`
	Data   []byte `yaml:"-"`
}

type PackageInfo struct {
	PackageID `yaml:",inline"`
	// Size of chaincode package
	Size int `yaml:"-"`
	// CreatedAt is a package creation date and time
	CreatedAt time.Time `yaml:"-"`
}
