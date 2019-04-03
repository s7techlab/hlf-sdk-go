package testdata

import (
	"path"
	"path/filepath"
	"runtime"
)

const (
	OrdererAddress = `orderer:7050`
	OrdererMspId   = `OrdererMSP`
)

var (
	OrdererMspPath = ``
)

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic(`failed to get runtime.Caller`)
	}

	OrdererMspPath = filepath.Join(path.Dir(file), `crypto-config`, `ordererOrganizations`, `example.com`, `orderers`, `orderer.example.com`, `msp`)
}
