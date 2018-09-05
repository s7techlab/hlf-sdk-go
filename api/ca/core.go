package ca

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
)

type Core interface {
	// Getting information about CA
	CAInfo(ctx context.Context) (*ResponseCAInfo, error)

	// Common operations over certificates
	Register(ctx context.Context, req RegistrationRequest) (string, error)
	Enroll(ctx context.Context, name, secret string, req *x509.CertificateRequest, opts ...EnrollOpt) (*x509.Certificate, interface{}, error)
	Revoke(ctx context.Context, req RevocationRequest) (*pkix.CertificateList, error)

	// Operations over identities
	IdentityList(ctx context.Context) ([]Identity, error)
	IdentityGet(ctx context.Context, enrollId string) (*Identity, error)

	// Operations over certificates
	CertificateList(ctx context.Context, opts ...CertificateListOpt) ([]*x509.Certificate, error)

	// Operations over affiliations
	// AffiliationList lists all affiliations and identities of identity affiliation
	AffiliationList(ctx context.Context, rootAffiliation ...string) ([]Identity, []Affiliation, error)
	AffiliationCreate(ctx context.Context, name string, opts ...AffiliationOpt) error
	AffiliationDelete(ctx context.Context, name string, opts ...AffiliationOpt) ([]Identity, []Affiliation, error)
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

type AffiliationOpt func(values *url.Values) error

func WithForce() AffiliationOpt {
	return func(values *url.Values) error {
		values.Set(`force`, `true`)
		return nil
	}
}
