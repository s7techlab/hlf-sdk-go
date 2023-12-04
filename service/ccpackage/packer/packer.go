package packer

import (
	"context"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
)

type (
	Packer interface {
		PackFromTar(ctx context.Context, spec *ccpackage.PackageSpec, tar []byte) (*ccpackage.Package, error)
		PackFromFiles(ctx context.Context, spec *ccpackage.PackageSpec, path string) (*ccpackage.Package, error)
	}

	PackFromTarRequest struct {
	}

	PackFromFileRequest struct {
	}
)
