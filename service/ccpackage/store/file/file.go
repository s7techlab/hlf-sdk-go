package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage/store"
)

type Params struct {
	RootPath string `mapstructure:"root_path"`
}

type Store struct {
	rootPath string
	perm     os.FileMode
}

func New(ps Params) *Store {
	return &Store{
		rootPath: ps.RootPath,
		perm:     0644,
	}
}

func (s *Store) filePath(id *ccpackage.PackageID) string {
	return filepath.Join(s.rootPath, store.ObjectKey(id))
}

func (s *Store) Put(ctx context.Context, req *ccpackage.PutPackageRequest) error {
	return os.WriteFile(s.filePath(req.Id), req.Data, s.perm)
}

// Get gets chaincode package info from storage.
func (s *Store) Get(_ context.Context, id *ccpackage.PackageID) (*ccpackage.PackageData, error) {
	fStat, err := os.Stat(s.filePath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, store.ErrPackageNotFound
		}
		return nil, fmt.Errorf("stat file: %w", err)
	}

	return &ccpackage.PackageData{
		Id:        id,
		Size:      fStat.Size(),
		CreatedAt: timestamppb.New(fStat.ModTime()),
	}, nil
}

// List gets stored chaincode packages' infos.
func (s *Store) List(ctx context.Context) ([]*ccpackage.Package, error) {
	fs, err := ioutil.ReadDir(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("read root path: %w", err)
	}

	var ps []*ccpackage.Package

	for _, f := range fs {

		if f.IsDir() {
			continue
		}

		name, version, fabricVersion, ok := store.ParseObjectKey(f.Name())
		if !ok {
			continue
		}

		ps = append(ps, &ccpackage.Package{
			Id: &ccpackage.PackageID{
				Name:          name,
				Version:       version,
				FabricVersion: fabricVersion,
			},
			Size:      f.Size(),
			CreatedAt: timestamppb.New(f.ModTime()),
		})
	}

	return ps, nil
}

// Fetch fetches chaincode package.
func (s *Store) Fetch(ctx context.Context, id *ccpackage.PackageID) (io.ReadCloser, error) {
	return os.Open(s.filePath(id))
}

func (s *Store) Close() error {
	return nil
}
