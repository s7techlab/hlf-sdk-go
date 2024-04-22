package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"

	"github.com/cloudflare/cfssl/signer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/client/ca"
)

const enrollEndpoint = `/api/v1/enroll`

func (c *Client) Enroll(ctx context.Context, name, secret string, req *x509.CertificateRequest, opts ...ca.EnrollOpt) (*x509.Certificate, interface{}, error) {
	var err error

	options := &ca.EnrollOpts{}
	for _, opt := range opts {
		if err = opt(options); err != nil {
			return nil, nil, errors.Wrap(err, `failed to apply option`)
		}
	}

	if options.PrivateKey == nil {
		if options.PrivateKey, err = c.crypto.NewPrivateKey(); err != nil {
			return nil, nil, errors.Wrap(err, `failed to generate private key`)
		}
	}

	if options.Profile == "" {
		options.Profile = ca.EnrollProfileDefault
	}

	// Add default signature algorithm if not defined
	if req.SignatureAlgorithm == x509.UnknownSignatureAlgorithm {
		req.SignatureAlgorithm = c.crypto.GetSignatureAlgorithm()
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, req, options.PrivateKey)
	if err != nil {
		return nil, options.PrivateKey, errors.Wrap(err, `failed to get certificate request`)
	}

	pemCsr := pem.EncodeToMemory(&pem.Block{Type: `CERTIFICATE REQUEST`, Bytes: csr})

	reqBytes, err := json.Marshal(signer.SignRequest{Request: string(pemCsr), Profile: string(options.Profile)})
	if err != nil {
		return nil, options.PrivateKey, errors.Wrap(err, `failed to marshal CSR request to JSON`)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.config.Host+enrollEndpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, options.PrivateKey, errors.Wrap(err, `failed to create http request`)
	}
	httpReq.SetBasicAuth(name, secret)

	resp, err := c.client.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, options.PrivateKey, errors.Wrap(err, `failed to send http request`)
	}

	var enrollResp ca.ResponseEnrollment

	if err = c.processResponse(resp, &enrollResp, http.StatusCreated); err != nil {
		return nil, options.PrivateKey, err
	}

	certDecoded, err := base64.StdEncoding.DecodeString(enrollResp.Cert)
	if err != nil {
		return nil, options.PrivateKey, errors.Wrap(err, `failed to decode base64 certificate`)
	}

	certBlock, _ := pem.Decode(certDecoded)
	if certBlock == nil {
		return nil, options.PrivateKey, errors.New(`failed to decode PEM block`)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, options.PrivateKey, errors.Wrap(err, `failed to parse certificate`)
	}

	return cert, options.PrivateKey, nil
}
