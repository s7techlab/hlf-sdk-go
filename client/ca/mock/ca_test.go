package mock_test

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/s7techlab/hlf-sdk-go/client/ca"
	"github.com/s7techlab/hlf-sdk-go/client/ca/mock"
)

var (
	//go:embed testdata/ca.pem
	cert []byte
	//go:embed testdata/ca-key.pem
	pk []byte

	client ca.Client

	ctx = context.Background()
)

func TestNewMockCaClientInvalidCert(t *testing.T) {
	var err error
	client, err = mock.New([]byte(`invalid`), []byte(`invalid_cert`))
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestNewMockCaClientValid(t *testing.T) {
	var err error
	client, err = mock.New(pk, cert)
	assert.NoError(t, err)
	assert.NotNil(t, client)

}

func TestCaClient_Enroll(t *testing.T) {
	req := &x509.CertificateRequest{Subject: pkix.Name{
		Country:       []string{`RU`},
		Organization:  []string{`my org`},
		StreetAddress: []string{`my address`},
		CommonName:    `org1`,
	}}

	certificate, privateKey, err := client.Enroll(ctx, ``, ``, req)
	assert.NoError(t, err)
	assert.NotNil(t, certificate)
	assert.NotNil(t, privateKey)

	assert.Equal(t, certificate.Subject.CommonName, req.Subject.CommonName)
}
