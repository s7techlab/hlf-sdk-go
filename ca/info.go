package ca

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/ca"
)

func (c *core) CAInfo() (*ca.ResponseCAInfo, error) {
	req, err := http.NewRequest(`GET`, c.config.Host+`/api/v1/cainfo`, nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create http request`)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to process http request`)
	}

	var caInfoResp ca.ResponseCAInfo

	if err = c.processResponse(resp, &caInfoResp, http.StatusOK); err != nil {
		return nil, err
	}

	return &caInfoResp, nil
}
