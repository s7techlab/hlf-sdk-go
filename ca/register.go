package ca

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/ca"
)

const regEndpoint = `/api/v1/register`

func (c *core) Register(req ca.RegistrationRequest) (string, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return ``, errors.Wrap(err, `failed to marshal request to JSON`)
	}

	authToken, err := c.createAuthToken(reqBytes)
	if err != nil {
		return ``, errors.Wrap(err, `failed to get auth token`)
	}

	httpReq, err := http.NewRequest(`POST`, c.config.Host+regEndpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return ``, errors.Wrap(err, `failed to create http request`)
	}

	httpReq.Header.Set(`Content-Type`, `application/json`)
	httpReq.Header.Set(`authorization`, authToken)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return ``, errors.Wrap(err, `failed to get response`)
	}

	var regResp ca.ResponseRegistration

	if err = c.processResponse(resp, &regResp, http.StatusCreated); err != nil {
		return ``, err
	}

	return regResp.Secret, nil
}
