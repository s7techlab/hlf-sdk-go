package camock

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	caapi "github.com/atomyze-ru/hlf-sdk-go/api/ca"
)

type (
	Cert struct {
		Cert       *x509.Certificate
		PrivateKey interface{}
	}

	CA struct {
		PK      interface{}
		Cert    *x509.Certificate
		CAChain string

		Enrolled map[string][]*x509.Certificate

		certCount      int64
		certCountMutex sync.Mutex
	}

	Opt func(*CA) error
)

func New(privateKey, cert []byte, opts ...Opt) (*CA, error) {
	var err error

	c := &CA{
		Enrolled: make(map[string][]*x509.Certificate),
	}

	if privateKey != nil {
		pkb, _ := pem.Decode(privateKey)
		if pkb == nil {
			return nil, errors.New(`failed to decode PEM pk`)
		}

		if c.PK, err = x509.ParsePKCS8PrivateKey(pkb.Bytes); err != nil {
			return nil, fmt.Errorf(`parse private key: %w`, err)
		}
	}

	if cert != nil {
		cb, _ := pem.Decode(cert)
		if cb == nil {
			return nil, fmt.Errorf(`failed decode PEM cert`)
		}

		if c.Cert, err = x509.ParseCertificate(cb.Bytes); err != nil {
			return nil, fmt.Errorf(`parse certificate: %w`, err)
		}
	}

	for _, opt := range opts {
		if err = opt(c); err != nil {
			return nil, err
		}
	}

	if c.CAChain == `` {
		c.CAChain = string(cert)
	}
	return c, nil
}

func (c *CA) CAInfo(ctx context.Context) (*caapi.ResponseCAInfo, error) {
	return &caapi.ResponseCAInfo{
		CAName:  "Mocked CA",
		CAChain: base64.StdEncoding.EncodeToString([]byte(c.CAChain)),
		Version: "",
	}, nil
}

func (c *CA) Register(ctx context.Context, req caapi.RegistrationRequest) (string, error) {
	return ``, nil
}

func (c *CA) Enroll(_ context.Context, name, _ string, req *x509.CertificateRequest, _ ...caapi.EnrollOpt) (*x509.Certificate, interface{}, error) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf(`generate private key: %w`, err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, c.templateFromCSR(req), c.Cert, pk.Public(), c.PK)
	if err != nil {
		return nil, nil, fmt.Errorf(`create certificate: %w`, err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, fmt.Errorf(`parse created cert: %w`, err)
	}

	if _, ok := c.Enrolled[name]; !ok {
		c.Enrolled[name] = []*x509.Certificate{}
	}

	c.Enrolled[name] = append(c.Enrolled[name], cert)

	return cert, pk, nil
}

func (c *CA) Revoke(ctx context.Context, req caapi.RevocationRequest) (*pkix.CertificateList, error) {
	panic("implement me")
}

func (c *CA) IdentityList(ctx context.Context) ([]caapi.Identity, error) {
	panic("implement me")
}

func (c *CA) IdentityGet(ctx context.Context, enrollId string) (*caapi.Identity, error) {
	panic("implement me")
}

func (c *CA) CertificateList(ctx context.Context, opts ...caapi.CertificateListOpt) ([]*x509.Certificate, error) {
	panic("implement me")
}

func (c *CA) AffiliationList(ctx context.Context, rootAffiliation ...string) ([]caapi.Identity, []caapi.Affiliation, error) {
	panic("implement me")
}

func (c *CA) AffiliationCreate(ctx context.Context, name string, opts ...caapi.AffiliationOpt) error {
	panic("implement me")
}

func (c *CA) AffiliationDelete(ctx context.Context, name string, opts ...caapi.AffiliationOpt) ([]caapi.Identity, []caapi.Affiliation, error) {
	panic("implement me")
}

func (c *CA) templateFromCSR(csr *x509.CertificateRequest) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(c.newSerialNumber()),
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}

func (c *CA) newSerialNumber() int64 {
	c.certCountMutex.Lock()
	defer c.certCountMutex.Unlock()
	c.certCount++
	return c.certCount
}

func WithCAChain(caCert string) func(*CA) error {
	return func(ca *CA) error {
		ca.CAChain = caCert
		return nil
	}
}

func MustNew(privateKey, cert []byte, opts ...Opt) *CA {
	ca, err := New(privateKey, cert, opts...)
	if err != nil {
		panic(err)
	}

	return ca
}
