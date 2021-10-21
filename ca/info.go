package ca

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api/ca"
)

func (c *core) CAInfo(ctx context.Context) (*ca.ResponseCAInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.config.Host+`/api/v1/cainfo`, nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create http request`)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, `failed to process http request`)
	}

	var caInfoResp ca.ResponseCAInfo

	if err = c.processResponse(resp, &caInfoResp, http.StatusOK); err != nil {
		return nil, err
	}

	return &caInfoResp, nil
}
