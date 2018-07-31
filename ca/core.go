package ca

import (
	"net/http"

	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/msp"
	mspPb "github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/ca"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/crypto"
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

func NewCore(mspId string, identity api.Identity, opts ...opt) (ca.Core, error) {
	var err error

	c := &core{mspId: mspId}

	// Applying user opts
	for _, opt := range opts {
		if err = opt(c); err != nil {
			return nil, errors.Wrap(err, `failed to apply option`)
		}
	}

	if c.config == nil {
		return nil, api.ErrEmptyConfig
	}

	if c.cs, err = crypto.GetSuite(c.config.Crypto.Type, c.config.Crypto.Options); err != nil {
		return nil, errors.Wrap(err, `failed to initialize crypto suite`)
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	c.identity = identity.GetSigningIdentity(c.cs)

	return c, nil
}
