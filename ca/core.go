package ca

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	mspPb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/api/ca"
	"github.com/s7techlab/hlf-sdk-go/v2/api/config"
	"github.com/s7techlab/hlf-sdk-go/v2/crypto"
	"github.com/s7techlab/hlf-sdk-go/v2/crypto/ecdsa"
)

type core struct {
	mspId    string
	cs       api.CryptoSuite
	config   *config.CAConfig
	client   *http.Client
	identity msp.SigningIdentity
}

func (c *core) createAuthToken(request []byte) (string, error) {
	id, err := c.identity.Serialize()
	if err != nil {
		return ``, errors.Wrap(err, `failed to get serialized identity`)
	}

	var serId mspPb.SerializedIdentity

	if err = proto.Unmarshal(id, &serId); err != nil {
		return ``, errors.Wrap(err, `failed to unmarshal serialized identity`)
	}

	baseCert := base64.StdEncoding.EncodeToString(serId.IdBytes)
	baseReq := base64.StdEncoding.EncodeToString(request)

	if signature, err := c.identity.Sign([]byte(baseReq + `.` + baseCert)); err != nil {
		return ``, errors.Wrap(err, `failed to sign data`)
	} else {
		return fmt.Sprintf("%s.%s", baseCert, base64.StdEncoding.EncodeToString(signature)), nil
	}
}

func (c *core) setAuthToken(req *http.Request, body []byte) error {
	if token, err := c.createAuthToken(body); err != nil {
		return errors.Wrap(err, `failed to create auth token`)
	} else {
		req.Header.Add(`Authorization`, token)
	}
	return nil
}

func (c *core) processResponse(resp *http.Response, out interface{}, expectedHTTPStatuses ...int) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `failed to read response body`)
	}

	if !c.expectedHTTPStatus(resp.StatusCode, expectedHTTPStatuses...) {
		return api.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
	}

	var caResp ca.Response
	if err = json.Unmarshal(body, &caResp); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON response`)
	}

	if !caResp.Success {
		return ca.ResponseError{Errors: caResp.Errors}
	}

	if err = json.Unmarshal(caResp.Result, out); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON`)
	}

	return nil
}

func (c *core) expectedHTTPStatus(status int, expected ...int) bool {
	for _, s := range expected {
		if s == status {
			return true
		}
	}
	return false
}

func NewCore(mspId string, identity api.Identity, opts ...opt) (ca.Core, error) {
	var err error

	c := &core{mspId: mspId}

	// Applying user opts
	for _, opt := range opts {
		if err = opt(c); err != nil {
			return nil, fmt.Errorf(`apply ca.core option: %w`, err)
		}
	}

	if c.config == nil {
		return nil, api.ErrEmptyConfig
	}

	if c.config.Crypto.Type == `` {
		c.config.Crypto = ecdsa.DefaultConfig
	}

	if c.cs, err = crypto.GetSuite(c.config.Crypto.Type, c.config.Crypto.Options); err != nil {
		return nil, fmt.Errorf(`initialize crypto suite: %w`, err)
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	c.identity = identity.GetSigningIdentity(c.cs)

	return c, nil
}
