package fetcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode/fetcher"
	"github.com/s7techlab/hlf-sdk-go/builder/chaincode/testdata"
)

func TestNewGoGitFetcher(t *testing.T) {
	f := fetcher.NewGit(fetcher.GitBasicAuth(testdata.User, testdata.Password))
	assert.NotNil(t, f)
}

func TestGoGitFetcher_Fetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	//for _, p := range testdata.Packages {
	//	code, err := f.Fetch(context.Background(), p.Repository, p.Version)
	//	assert.NoError(t, err)
	//	assert.NotNil(t, code)
	//	assert.NoError(t, expectTar(code))
	//}
}
