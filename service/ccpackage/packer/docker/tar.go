package docker

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
)

var ErrFileNotFound = errors.New(`file not found`)

func UnTarFirstFile(r io.Reader) (*bytes.Buffer, error) {
	tr := tar.NewReader(r)
	file := new(bytes.Buffer)

	header, err := tr.Next()

	switch {
	case header == nil:
		return nil, nil

	case err != nil:
		return nil, err
	}

	switch header.Typeflag {
	case tar.TypeReg:
		if _, err := io.CopyN(file, tr, header.Size); err != nil {
			return nil, err
		}
		return file, nil

	default:
		return nil, ErrFileNotFound
	}
}
