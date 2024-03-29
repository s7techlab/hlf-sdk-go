package fetcher

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"go.uber.org/zap"
)

type File struct {
	Logger *zap.Logger
}

func NewFile(l *zap.Logger) *File {
	return &File{
		Logger: l,
	}
}

func (f *File) Fetch(ctx context.Context, path, _ string) (code []byte, err error) {
	bf := new(bytes.Buffer)
	tw := tar.NewWriter(bf)

	defer func() {
		twErr := tw.Close()
		if err == nil && twErr != nil {
			err = fmt.Errorf("close tar writer: %w", err)
		}
	}()

	path = strings.TrimPrefix(path, FileProtocolPrefix)

	fs := osfs.New(path)

	f.Logger.Debug(`adding path to tar`, zap.String(`path`, path))

	if err = AddFileToTar(tw, `/`, fs); err != nil {
		err = fmt.Errorf("fetch filepath=%s to tar: %w", path, err)
		return
	}

	return bf.Bytes(), nil
}
