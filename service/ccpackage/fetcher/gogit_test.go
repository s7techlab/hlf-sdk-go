package fetcher_test

//import (
//	"archive/tar"
//	"bytes"
//	"context"
//	"io"
//	"testing"
//
//	"github.com/stretchr/testify/assert"

//
//var (
//	f fetcher.Fetcher
//)
//
//func TestNewGoGitFetcher(t *testing.T) {
//	f = fetcher.NewGit(fetcher.GitBasicAuth(testdata.DeployUser, testdata.DeployPassword))
//	assert.NotNil(t, f)
//}
//
//func TestGoGitFetcher_Fetch(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping test in short mode.")
//	}
//
//	for _, p := range testdata.Packages {
//		code, err := f.Fetch(context.Background(), p.Repository, p.Version)
//		assert.NoError(t, err)
//		assert.NotNil(t, code)
//		assert.NoError(t, expectTar(code))
//	}
//}
//
