package fetcher_test

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage/fetcher"
)

func TestFileFetcher_Fetch(t *testing.T) {
	f := fetcher.NewFile(zap.NewNop())
	assert.NotNil(t, f)

	abs, err := filepath.Abs(`../`)
	assert.NoError(t, err)

	code, err := f.Fetch(context.Background(), abs+`/`, `some ver`)
	assert.NoError(t, err)
	assert.NotNil(t, code)
	assert.NoError(t, expectTar(code))
}

func expectTar(b []byte) error {
	r := bytes.NewReader(b)
	tr := tar.NewReader(r)
	for {
		_, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
