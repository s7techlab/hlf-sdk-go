package ca

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-ca/api"
)

type Core interface {
	Register(req api.RegistrationRequest) (string, error)
	Enroll(name, secret string, req *x509.CertificateRequest, opts ...EnrollOpt) (*x509.Certificate, interface{}, error)
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
