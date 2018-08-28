package ca

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
)

type Core interface {
	// Getting information about CA
	CAInfo() (*ResponseCAInfo, error)

	// Common operations over certificates
	Register(req RegistrationRequest) (string, error)
	Enroll(name, secret string, req *x509.CertificateRequest, opts ...EnrollOpt) (*x509.Certificate, interface{}, error)
	Revoke(ctx context.Context, req RevocationRequest) (*pkix.CertificateList, error)

	// Operations over identities
	IdentityList(ctx context.Context) ([]Identity, error)
	IdentityGet(ctx context.Context, enrollId string) (*Identity, error)

	// Operations over certificates
	CertificateList(ctx context.Context, opts ...CertificateListOpt) ([]*x509.Certificate, error)
}

type EnrollOpts struct {
	PrivateKey interface{}
}

type EnrollOpt func(opts *EnrollOpts) error

func WithEnrollPrivateKey(privateKey interface{}) EnrollOpt {
	return func(opts *EnrollOpts) error {
		opts.PrivateKey = privateKey
		return nil
	}
}

type CertificateListOpt func(values *url.Values) error

func WithEnrollId(enrollId string) CertificateListOpt {
	return func(values *url.Values) error {
		values.Add(`id`, enrollId)
		return nil
	}
}
