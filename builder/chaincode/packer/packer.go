package packer

import (
	"context"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
)

type Packer interface {
	// Pack creates binary package of chaincode by module path
	Pack(ctx context.Context, pkg *chaincode.Package) error
}
