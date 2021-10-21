package ca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api/ca"
)

const (
	endpointAffiliationList   = "%s/api/v1/affiliations%s"
	endpointAffiliationCreate = "%s/api/v1/affiliations%s"
	endpointAffiliationDelete = "%s/api/v1/affiliations/%s"
)

func (c *core) AffiliationList(ctx context.Context, rootAffiliation ...string) ([]ca.Identity, []ca.Affiliation, error) {
	var reqUrl string

	if len(rootAffiliation) == 1 {
		reqUrl = fmt.Sprintf(endpointAffiliationList, c.config.Host, `/`+rootAffiliation[0])
	} else {
		reqUrl = fmt.Sprintf(endpointAffiliationList, c.config.Host, ``)
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to create request`)
	}

	if err = c.setAuthToken(req, nil); err != nil {
		return nil, nil, errors.Wrap(err, `failed to set auth token`)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to process request`)
	}

	var affiliationResponse ca.ResponseAffiliationList

	if err = c.processResponse(resp, &affiliationResponse, http.StatusOK, http.StatusCreated); err != nil {
		return nil, nil, err
	}

	return affiliationResponse.Identities, affiliationResponse.Affiliations, nil
}

func (c *core) AffiliationCreate(ctx context.Context, name string, opts ...ca.AffiliationOpt) error {
	var (
		reqUrl string
		err    error
	)
	u := url.Values{}

	for _, opt := range opts {
		if err = opt(&u); err != nil {
			return errors.Wrap(err, `failed to apply option`)
		}
	}

	if v := u.Encode(); v == `` {
		reqUrl = fmt.Sprintf(endpointAffiliationCreate, c.config.Host, ``)
	} else {
		reqUrl = fmt.Sprintf(endpointAffiliationCreate, c.config.Host, `?`+v)
	}

	reqBytes, err := json.Marshal(ca.AddAffiliationRequest{Name: name})
	if err != nil {
		return errors.Wrap(err, `failed to marshal JSON request`)
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return errors.Wrap(err, `failed to create request`)
	}

	if err = c.setAuthToken(req, reqBytes); err != nil {
		return errors.Wrap(err, `failed to set auth token`)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, `failed to process request`)
	}

	var affiliationCreateResponse ca.ResponseAffiliationCreate

	if err = c.processResponse(resp, &affiliationCreateResponse, http.StatusCreated); err != nil {
		return err
	}

	return nil
}

func (c *core) AffiliationDelete(ctx context.Context, name string, opts ...ca.AffiliationOpt) ([]ca.Identity, []ca.Affiliation, error) {
	var (
		reqUrl string
		err    error
	)

	u := url.Values{}

	for _, opt := range opts {
		if err = opt(&u); err != nil {
			return nil, nil, errors.Wrap(err, `failed to apply option`)
		}
	}

	if v := u.Encode(); v == `` {
		reqUrl = fmt.Sprintf(endpointAffiliationDelete, c.config.Host, name)
	} else {
		reqUrl = fmt.Sprintf(endpointAffiliationDelete, c.config.Host, name+`?`+v)
	}

	req, err := http.NewRequest(http.MethodDelete, reqUrl, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to create request`)
	}

	if err = c.setAuthToken(req, nil); err != nil {
		return nil, nil, errors.Wrap(err, `failed to set auth token`)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to process request`)
	}

	var affiliationDeleteResponse ca.ResponseAffiliationDelete

	if err = c.processResponse(resp, &affiliationDeleteResponse, http.StatusOK); err != nil {
		return nil, nil, err
	}

	return affiliationDeleteResponse.Identities, affiliationDeleteResponse.Affiliations, nil
}
