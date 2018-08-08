package ca

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/mitchellh/mapstructure"
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected http status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, `failed to read response body`)
	}

	var caResponse ca.Response
	if err = json.Unmarshal(body, &caResponse); err != nil {
		return nil, errors.Wrap(err, `failed to parse response body`)
	}

	if caResponse.Success != true {
		return nil, &ca.ResponseError{Errors: caResponse.Errors}
	}
	switch result := caResponse.Result.(type) {
	case map[string]interface{}:
		var caInfoResp ca.ResponseCAInfo
		if err = mapstructure.Decode(result, &caInfoResp); err != nil {
			return nil, errors.Wrap(err, `failed to decode response result`)
		}
		return &caInfoResp, nil
	default:
		return nil, errors.Errorf("unexpected response type:%s", reflect.ValueOf(caResponse.Result).Type().String())
	}
}
