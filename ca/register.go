package ca

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/cloudflare/cfssl/api"
	caApi "github.com/hyperledger/fabric-ca/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	sdkApi "github.com/s7techlab/hlf-sdk-go/api"
)

const regEndpoint = `/api/v1/register`

func (c *core) Register(req caApi.RegistrationRequest) (string, error) {
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

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ``, errors.Wrap(err, `failed to read response body`)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		res := api.Response{}
		if err = json.Unmarshal(body, &res); err != nil {
			return ``, errors.Wrap(err, `failed to unmarshal JSON response`)
		}

		if res.Success != true {
			return ``, &sdkApi.CAResponseError{Errors: res.Errors, Messages: res.Messages}
		}

		switch result := res.Result.(type) {
		case map[string]interface{}:
			var regResp caApi.RegistrationResponse
			if err = mapstructure.Decode(result, &regResp); err != nil {
				return ``, errors.Wrap(err, `failed to decode CA response`)
			}
			return regResp.Secret, nil
		default:
			return ``, errors.Errorf("unexpected response type:%s", reflect.ValueOf(res.Result).Type().String())
		}
	} else {
		return ``, errors.Errorf("http response error: %d %s", resp.StatusCode, string(body))
	}
}
