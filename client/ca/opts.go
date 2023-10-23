package ca

import (
	"net/url"
)

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
