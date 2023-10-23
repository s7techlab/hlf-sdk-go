package ca

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
)

type Client interface {
	// CAInfo Getting information about CA
	CAInfo(ctx context.Context) (*ResponseCAInfo, error)

	Register(ctx context.Context, req RegistrationRequest) (string, error)
	Enroll(ctx context.Context, name, secret string, req *x509.CertificateRequest, opts ...EnrollOpt) (
		*x509.Certificate, interface{}, error)
	Revoke(ctx context.Context, req RevocationRequest) (*pkix.CertificateList, error)

	IdentityList(ctx context.Context) ([]Identity, error)
	IdentityGet(ctx context.Context, enrollId string) (*Identity, error)

	CertificateList(ctx context.Context, opts ...CertificateListOpt) ([]*x509.Certificate, error)

	// AffiliationList lists all affiliations and identities of identity affiliation
	AffiliationList(ctx context.Context, rootAffiliation ...string) ([]Identity, []Affiliation, error)
	AffiliationCreate(ctx context.Context, name string, opts ...AffiliationOpt) error
	AffiliationDelete(ctx context.Context, name string, opts ...AffiliationOpt) ([]Identity, []Affiliation, error)
}
