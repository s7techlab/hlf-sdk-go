package ca

import (
	"context"
	"crypto/x509"
)

type Core interface {
	CAInfo() (*ResponseCAInfo, error)
	Register(req RegistrationRequest) (string, error)
	Enroll(name, secret string, req *x509.CertificateRequest, opts ...EnrollOpt) (*x509.Certificate, interface{}, error)

	IdentityList(ctx context.Context) ([]Identity, error)
	IdentityGet(ctx context.Context, enrollId string) (*Identity, error)
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
