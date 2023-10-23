package http

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

	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client"
	"github.com/s7techlab/hlf-sdk-go/client/ca"
	clienterrors "github.com/s7techlab/hlf-sdk-go/client/errors"
	"github.com/s7techlab/hlf-sdk-go/crypto"
)

type Client struct {
	crypto crypto.Suite
	config *config.CAConfig
	client *http.Client
	signer msp.SigningIdentity
}

func New(signer msp.SigningIdentity, opts ...Opt) (*Client, error) {
	var err error

	c := &Client{
		signer: signer,
	}

	for _, opt := range opts {
		if err = opt(c); err != nil {
			return nil, fmt.Errorf(`apply ca.Client option: %w`, err)
		}
	}

	if c.config == nil {
		return nil, client.ErrEmptyConfig
	}

	if c.crypto == nil {
		c.crypto, err = crypto.NewSuiteByConfig(c.config.Crypto, true)
		if err != nil {
			return nil, err
		}
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	c.signer = signer

	return c, nil
}

func (c *Client) createAuthToken(request []byte) (string, error) {
	id, err := c.signer.Serialize()
	if err != nil {
		return ``, errors.Wrap(err, `failed to get serialized signer`)
	}

	var serId mspPb.SerializedIdentity

	if err = proto.Unmarshal(id, &serId); err != nil {
		return ``, errors.Wrap(err, `failed to unmarshal serialized signer`)
	}

	baseCert := base64.StdEncoding.EncodeToString(serId.IdBytes)
	baseReq := base64.StdEncoding.EncodeToString(request)

	if signature, err := c.signer.Sign([]byte(baseReq + `.` + baseCert)); err != nil {
		return ``, errors.Wrap(err, `failed to sign data`)
	} else {
		return fmt.Sprintf("%s.%s", baseCert, base64.StdEncoding.EncodeToString(signature)), nil
	}
}

func (c *Client) setAuthToken(req *http.Request, body []byte) error {
	if token, err := c.createAuthToken(body); err != nil {
		return errors.Wrap(err, `failed to create auth token`)
	} else {
		req.Header.Add(`Authorization`, token)
	}
	return nil
}

func (c *Client) processResponse(resp *http.Response, out interface{}, expectedHTTPStatuses ...int) error {
	defer func() { _ = resp.Body.Close() }()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `failed to read response body`)
	}

	if !c.expectedHTTPStatus(resp.StatusCode, expectedHTTPStatuses...) {
		return clienterrors.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
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

func (c *Client) expectedHTTPStatus(status int, expected ...int) bool {
	for _, s := range expected {
		if s == status {
			return true
		}
	}
	return false
}
