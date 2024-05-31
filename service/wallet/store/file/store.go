package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/s7techlab/hlf-sdk-go/proto/wallet"
	wallet2 "github.com/s7techlab/hlf-sdk-go/service/wallet"
)

var (
	ErrDirectoryNotExists = errors.New(`directory not exists`)
)

type (
	FilesystemStore struct {
		baseDir string
	}
)

func New(baseDir string) (*FilesystemStore, error) {
	if baseDir == `` {
		return nil, fmt.Errorf(`dir= : %w`, ErrDirectoryNotExists)
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf(`dir= %s: %w`, baseDir, ErrDirectoryNotExists)
	}

	if baseDir[len(baseDir)-1:] != `/` {
		baseDir = baseDir + `/`
	}

	return &FilesystemStore{baseDir: baseDir}, nil
}

func (f *FilesystemStore) labelToFilename(label string) string {
	return f.baseDir + label + `.json`
}

func (f *FilesystemStore) filenameToLabel(filename string) string {
	if len(filename) < 6 {
		return ``
	}

	if filename[len(filename)-5:] != `.json` {
		return ``
	}
	return filename[0 : len(filename)-5]
}

func (f *FilesystemStore) Get(label string) (*wallet.IdentityInWallet, error) {
	if _, err := os.Stat(f.labelToFilename(label)); os.IsNotExist(err) {
		return nil, wallet2.ErrIdentityNotFound
	}

	bb, err := os.ReadFile(f.labelToFilename(label))
	if err != nil {
		return nil, err
	}

	identity := new(wallet.IdentityInWallet)

	if err = json.Unmarshal(bb, identity); err != nil {
		return nil, err
	}

	return identity, nil
}

func (f *FilesystemStore) Set(identity *wallet.IdentityInWallet) error {
	bb, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	return os.WriteFile(f.labelToFilename(identity.Label), bb, 0600)
}

func (f *FilesystemStore) List() ([]string, error) {
	var labels []string

	files, err := os.ReadDir(f.baseDir)
	if err != nil {
		return nil, fmt.Errorf(`reading %s: %w`, f.baseDir, err)
	}

	for _, file := range files {
		if label := f.filenameToLabel(file.Name()); label != `` {
			labels = append(labels, label)
		}
	}
	return labels, nil
}

func (f *FilesystemStore) Delete(label string) error {
	if _, err := os.Stat(f.labelToFilename(label)); os.IsNotExist(err) {
		return wallet2.ErrIdentityNotFound
	}

	return os.Remove(f.labelToFilename(label))
}
