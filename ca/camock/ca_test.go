package camock_test

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/s7techlab/hlf-sdk-go/api/ca"
	"github.com/s7techlab/hlf-sdk-go/ca/camock"
)

var (
	//go:embed testdata/ca.pem
	cert []byte
	//go:embed testdata/ca-key.pem
	pk []byte

	core ca.Core

	ctx = context.Background()
)

func TestNewMockCaClientInvalidCert(t *testing.T) {
	var err error
	core, err = camock.New([]byte(`invalid`), []byte(`invalid_cert`))
	assert.Nil(t, core)
	assert.Error(t, err)
}

func TestNewMockCaClientValid(t *testing.T) {
	var err error
	core, err = camock.New(pk, cert)
	assert.NoError(t, err)
	assert.NotNil(t, core)

}

func TestCaClient_Enroll(t *testing.T) {
	req := &x509.CertificateRequest{Subject: pkix.Name{
		Country:       []string{`RU`},
		Organization:  []string{`my org`},
		StreetAddress: []string{`my address`},
		CommonName:    `org1`,
	}}

	certificate, privateKey, err := core.Enroll(ctx, ``, ``, req)
	assert.NoError(t, err)
	assert.NotNil(t, certificate)
	assert.NotNil(t, privateKey)

	assert.Equal(t, certificate.Subject.CommonName, req.Subject.CommonName)
}
