package ecdsa

import (
	"testing"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/stretchr/testify/assert"
)

var (
	suite api.CryptoSuite
	err   error

	correctEcdsaOpts = config.CryptoSuiteOpts{`curve`: `P256`, `signatureAlgorithm`: `SHA256`, `hash`: `SHA2-256`}
)

func TestEcdsaSuite_Initialize(t *testing.T) {
	s := new(ecdsaSuite)
	suite, err = s.Initialize(correctEcdsaOpts)
	assert.NoError(t, err)
}

func TestEcdsaSuite_Sign(t *testing.T) {
	assert.NotNil(t, suite)
}
