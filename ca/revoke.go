package ca

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api/ca"
)

const (
	endpointRevoke = "%s/api/v1/revoke"
)

func (c *core) Revoke(ctx context.Context, req ca.RevocationRequest) (*pkix.CertificateList, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to marshal JSON request`)
	}

	httpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(endpointRevoke, c.config.Host), bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, errors.Wrap(err, `failed to create request`)
	}

	if err = c.setAuthToken(httpReq, reqBytes); err != nil {
		return nil, errors.Wrap(err, `failed to set auth token`)
	}

	resp, err := c.client.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, `failed to process request`)
	}

	var revokeResponse ca.ResponseRevoke

	if err = c.processResponse(resp, &revokeResponse, http.StatusOK); err != nil {
		return nil, err
	}

	if crl, err := x509.ParseCRL(revokeResponse.CRL); err != nil {
		return nil, errors.Wrap(err, `failed to parse CRL`)
	} else {
		return crl, nil
	}
}
