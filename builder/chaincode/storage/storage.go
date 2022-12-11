package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
)

type Storage interface {
	// Put saves chaincode package into storage.
	Put(ctx context.Context, pkg chaincode.Package) error
	// Get gets chaincode package info from storage.
	Get(ctx context.Context, id chaincode.PackageID) (
		chaincode.PackageInfo, error)
	// List gets stored chaincode packages' infos.
	List(ctx context.Context) ([]chaincode.PackageInfo, error)
	// Fetch fetches chaincode package.
	Fetch(ctx context.Context, id chaincode.PackageID) (io.ReadCloser, error)
	// Close closes storage
	Close() error
}

var ErrPackageNotFound = fmt.Errorf("package not found")

var objectNameRegexp = regexp.MustCompile(`^([^_\s]+)_([^_\s]+)_([^_\s]+)\.pkg$`)

func ParseObjectName(s string) (
	name string,
	version string,
	fabricVersion chaincode.FabricVersion,
	ok bool,
) {
	ss := objectNameRegexp.FindStringSubmatch(s)
	if len(ss) != 4 {
		return "", "", "", false
	}
	return ss[1], ss[2], chaincode.FabricVersion(ss[3]), true
}

func ObjectName(id chaincode.PackageID) string {
	return fmt.Sprintf("%s_%s_%s.pkg", id.Name, id.Version, id.FabricVersion)
}
