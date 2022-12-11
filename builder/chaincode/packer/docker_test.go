package packer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
	"github.com/s7techlab/hlf-sdk-go/builder/chaincode/packer"
)

var (
	dockerPacker *packer.Docker

	ctx    = context.Background()
	log, _ = zap.NewDevelopment()
)

func TestNewDocker(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	var err error
	dockerPacker, err = packer.New(chaincode.FabricV1, log)
	assert.NoError(t, err)
	assert.NotNil(t, dockerPacker)
}

func TestD_Pack(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	//for _, pkg := range testdata.Packages {
	//
	//	code, err := f.Fetch(ctx, pkg.Repository, pkg.Version)
	//	assert.NoError(t, err)
	//
	//	pkg.Source = code
	//	err = cli.Pack(ctx, &pkg)
	//	assert.NoError(t, err)
	//	assert.NotNil(t, pkg.Data)
	//}
}
