package ca

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/ca"
)

const (
	endpointIdentityList = `/api/v1/identities`
	endpointIdentityGet  = "/api/v1/identities/%s"
)

func (c *core) IdentityList(ctx context.Context) ([]ca.Identity, error) {
	req, err := http.NewRequest(http.MethodGet, endpointIdentityList, nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create request`)
	}

	req = req.WithContext(ctx)

	if err = c.setAuthToken(req, nil); err != nil {
		return nil, errors.Wrap(err, `failed to set auth token`)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to process request`)
	}

	var identityListResp ca.ResponseIdentityList

	if err = c.processResponse(resp, &identityListResp, http.StatusOK); err != nil {
		return nil, err
	}

	return identityListResp.Identities, nil
}

func (c *core) IdentityGet(ctx context.Context, enrollId string) (*ca.Identity, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(endpointIdentityGet, enrollId), nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create request`)
	}

	req = req.WithContext(ctx)

	if err = c.setAuthToken(req, nil); err != nil {
		return nil, errors.Wrap(err, `failed to set auth token`)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to process request`)
	}

	var identity ca.Identity

	if c.processResponse(resp, &identity, http.StatusOK); err != nil {
		return nil, err
	}

	return &identity, nil
}
