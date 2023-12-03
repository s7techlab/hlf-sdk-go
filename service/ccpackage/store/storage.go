package store

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
)

var (
	ErrPackageNotFound = fmt.Errorf("package not found")

	objectNameRegexp = regexp.MustCompile(`^([^_\s]+)_([^_\s]+)_([^_\s]+)\.pkg$`)
)

const (
	TypeS3     = "s3"
	TypeFile   = "file"
	TypeMemory = "memory"
)

type (
	Storage interface {
		// Put saves chaincode package into storage.
		Put(context.Context, *ccpackage.PutPackageRequest) error
		// Get gets chaincode package info from storage.
		Get(context.Context, *ccpackage.PackageID) (*ccpackage.PackageData, error)
		// List gets stored chaincode packages' infos.
		List(context.Context) ([]*ccpackage.Package, error)
		// Fetch fetches chaincode package.
		Fetch(context.Context, *ccpackage.PackageID) (io.ReadCloser, error)
		// Close closes storage
		Close() error
	}
)

func ParseObjectKey(s string) (
	name string,
	version string,
	fabricVersion ccpackage.FabricVersion,
	ok bool,
) {
	ss := objectNameRegexp.FindStringSubmatch(s)
	if len(ss) != 4 {
		return "", "", ccpackage.FabricVersion_FABRIC_VERSION_UNSPECIFIED, false
	}
	return ss[1], ss[2], ccpackage.FabricVersion(ccpackage.FabricVersion_value[ss[3]]), true
}

func ObjectKey(id *ccpackage.PackageID) string {
	return fmt.Sprintf("%s_%s_%s.pkg", id.Name, id.Version, id.FabricVersion)
}
