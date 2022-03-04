package ca

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api/ca"
)

const endpointCertificateList = "%s/api/v1/certificates%s"

func (c *core) CertificateList(ctx context.Context, opts ...ca.CertificateListOpt) ([]*x509.Certificate, error) {
	var (
		reqUrl string
		err    error
	)

	u := url.Values{}
	for _, opt := range opts {
		if err = opt(&u); err != nil {
			return nil, errors.Wrap(err, `failed to apply opt`)
		}
	}

	if v := u.Encode(); v == `` {
		reqUrl = fmt.Sprintf(endpointCertificateList, c.config.Host, ``)
	} else {
		reqUrl = fmt.Sprintf(endpointCertificateList, c.config.Host, `?`+v)
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create request`)
	}

	if err = c.setAuthToken(req, nil); err != nil {
		return nil, errors.Wrap(err, `failed to set authorization token`)
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to process request`)
	}

	var certResponse ca.ResponseCertificateList

	if err = c.processResponse(resp, &certResponse, http.StatusOK); err != nil {
		return nil, err
	}

	certs := make([]*x509.Certificate, len(certResponse.Certs))
	for i, v := range certResponse.Certs {
		b, _ := pem.Decode([]byte(v.PEM))
		if b == nil {
			return nil, errors.Errorf("failed to parse PEM block: %s", v)
		}
		if cert, err := x509.ParseCertificate(b.Bytes); err != nil {
			return nil, errors.Wrap(err, `failed to parse certificate`)
		} else {
			certs[i] = cert
		}
	}

	return certs, nil
}
