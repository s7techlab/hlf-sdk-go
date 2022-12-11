package camock_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/s7techlab/hlf-sdk-go/api/ca"
	"github.com/s7techlab/hlf-sdk-go/ca/camock"
)

var (
	cert = []byte(`-----BEGIN CERTIFICATE-----
MIICNzCCAd2gAwIBAgIUAxSYho+82wKuMwOR5ddhShil4UswCgYIKoZIzj0EAwIw
aTELMAkGA1UEBhMCUlUxDzANBgNVBAgTBk1vc2NvdzEPMA0GA1UEBxMGTW9zY293
MQswCQYDVQQKEwJTNzEQMA4GA1UECxMHVGVjaGxhYjEZMBcGA1UEAxMQZmFicmlj
LWNhLXNlcnZlcjAeFw0xODEwMTgxNDQ2MDBaFw0xOTEwMTgxNDUxMDBaMF0xCzAJ
BgNVBAYTAlVTMRcwFQYDVQQIEw5Ob3J0aCBDYXJvbGluYTEUMBIGA1UEChMLSHlw
ZXJsZWRnZXIxDzANBgNVBAsTBmNsaWVudDEOMAwGA1UEAxMFYWRtaW4wWTATBgcq
hkjOPQIBBggqhkjOPQMBBwNCAATSQiNUgFmdoUqP4tvmlKHt7QnUPbgpFqtEj3A/
r2yosevN4WpCSUj5gZVP+ZuUbUYRU9QfrdCUvYikMecgWSSDo28wbTAOBgNVHQ8B
Af8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUAcFuNkSOUXRq8RS6E8y8
YF+NeX8wHwYDVR0jBBgwFoAUFpTSOrkiHtXGsaZ6gU+V/wkvCawwDQYDVR0RBAYw
BIICY2EwCgYIKoZIzj0EAwIDSAAwRQIhALojGU0m0DvZqld3OHq3Dh4FsFM7PXDc
wZoFWUBCx7OMAiAOuVBSVZ5kYPiomu73SGI/d/gYKUjsISN7zMlt9CFC4Q==
-----END CERTIFICATE-----`)

	pk = []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgP8pqCjD0iCYShYYs
FjRz4WeE//m4SJKz3kn08RWH3UehRANCAAQIIOuPgnO3UvyI6my2E8wqdvj78Vfv
QX91VDRObDEMCV6gqbt1xQDc/kA3rJnB9vUNeJFRI4TVNdan365Atp+I
-----END PRIVATE KEY-----`)

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
	assert.NotNil(t, core)
	assert.NoError(t, err)
}

func TestCaClient_Enroll(t *testing.T) {
	//req := &x509.CertificateRequest{Subject: pkix.Name{
	//	Country:       []string{`RU`},
	//	Organization:  []string{`my org`},
	//	StreetAddress: []string{`my address`},
	//	CommonName:    `org1`,
	//}}
	//
	//certificate, privateKey, err := core.Enroll(ctx, ``, ``, req)
	//assert.NoError(t, err)
	//assert.NotNil(t, certificate)
	//assert.NotNil(t, privateKey)
	//
	//assert.Equal(t, certificate.Subject.CommonName, req.Subject.CommonName)
}
