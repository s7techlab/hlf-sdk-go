package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
)

type FileParams struct {
	RootPath string `mapstructure:"root_path"`
}

type fileStorage struct {
	rootPath string
	perm     os.FileMode
}

func NewFile(ps FileParams) *fileStorage {
	return &fileStorage{
		rootPath: ps.RootPath,
		perm:     0644,
	}
}

func (s *fileStorage) filePath(id chaincode.PackageID) string {
	return filepath.Join(s.rootPath, ObjectName(id))
}

func (s *fileStorage) Put(ctx context.Context, pkg chaincode.Package) error {
	return ioutil.WriteFile(s.filePath(pkg.PackageID),
		pkg.Data, s.perm)
}

// Get gets chaincode package info from storage.
func (s *fileStorage) Get(ctx context.Context, id chaincode.PackageID) (
	chaincode.PackageInfo, error) {

	fStat, err := os.Stat(s.filePath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return chaincode.PackageInfo{}, ErrPackageNotFound
		}
		return chaincode.PackageInfo{}, fmt.Errorf("stat file: %w", err)
	}

	return chaincode.PackageInfo{
		PackageID: id,
		Size:      int(fStat.Size()),
		CreatedAt: fStat.ModTime(),
	}, nil
}

// List gets stored chaincode packages' infos.
func (s *fileStorage) List(ctx context.Context) (
	[]chaincode.PackageInfo, error) {

	fs, err := ioutil.ReadDir(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("read root path: %w", err)
	}

	var ps []chaincode.PackageInfo

	for _, f := range fs {

		if f.IsDir() {
			continue
		}

		name, version, fabricVersion, ok := ParseObjectName(f.Name())
		if !ok {
			continue
		}

		ps = append(ps, chaincode.PackageInfo{
			PackageID: chaincode.PackageID{
				Name:          name,
				Version:       version,
				FabricVersion: fabricVersion,
			},
			Size:      int(f.Size()),
			CreatedAt: f.ModTime(),
		})
	}

	return ps, nil
}

// Fetch fetches chaincode package.
func (s *fileStorage) Fetch(ctx context.Context, id chaincode.PackageID) (
	io.ReadCloser, error) {
	return os.Open(s.filePath(id))
}

func (s *fileStorage) Close() error {
	return nil
}
