package ca

import (
	"net/url"
)

type EnrollProfile string

const (
	// EnrollProfileDefault asks Fabric CA for certificate used for signing
	EnrollProfileDefault EnrollProfile = ""
	// EnrollProfileTls asks Fabric CA for certificate used for TLS communication
	EnrollProfileTls EnrollProfile = "tls"
)

type EnrollOpts struct {
	PrivateKey interface{}
	Profile    EnrollProfile
}

type EnrollOpt func(opts *EnrollOpts) error

// WithEnrollPrivateKey allows to use previously created private key
func WithEnrollPrivateKey(privateKey interface{}) EnrollOpt {
	return func(opts *EnrollOpts) error {
		opts.PrivateKey = privateKey
		return nil
	}
}

// WithEnrollProfile allows to require profile of enrolled certificate
func WithEnrollProfile(profile EnrollProfile) EnrollOpt {
	return func(opts *EnrollOpts) error {
		opts.Profile = profile
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
