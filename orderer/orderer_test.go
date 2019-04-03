package orderer

import (
	"context"
	"testing"
	"time"

	"github.com/s7techlab/hlf-sdk-go/testdata"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/s7techlab/hlf-sdk-go/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	ord api.Orderer

	sampleOrdererConfig = config.ConnectionConfig{
		Host: testdata.OrdererAddress,
		Tls: config.TlsConfig{
			Enabled: false,
		},
		Timeout: config.Duration{Duration: 5 * time.Second},
	}
	log, _ = zap.NewProduction()

	id msp.SigningIdentity

	cs api.CryptoSuite

	// TODO: not a best practic make context for all test pkgs
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
)

func TestNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var err error
	ord, err = New(sampleOrdererConfig, log)
	assert.NoError(t, err)
	assert.NotNil(t, ord)
}
func TestOrderer_Deliver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Making SeekEnvelope to seek specified block
	startPos, endPos := api.SeekNewest()()
	env, err := util.SeekEnvelope(`testchainid`, startPos, endPos, id)
	assert.NoError(t, err)
	assert.NotNil(t, env)

	// Deliver block using SeekEnvelope
	block, err := ord.Deliver(ctx, env)
	assert.NoError(t, err)
	assert.NotNil(t, block)

	// Making seek envelope with block with config
	blockId, err := utils.GetLastConfigIndexFromBlock(block)
	assert.NoError(t, err)
	startPos, endPos = api.SeekSingle(blockId)()
	env, err = util.SeekEnvelope(`testchainid`, startPos, endPos, id)
	assert.NoError(t, err)
	assert.NotNil(t, env)

	// Deliver block with config
	block, err = ord.Deliver(ctx, env)
	assert.NoError(t, err)
	assert.NotNil(t, block)
}

func init() {
	// Initialize ECDSA crypto suite
	var err error
	cs, err = crypto.GetSuite(ecdsa.Module, config.CryptoSuiteOpts{`curve`: `P256`, `signatureAlgorithm`: `SHA256`, `hash`: `SHA2-256`})
	if err != nil {
		panic(err)
	}

	// Initializing signing identity
	mspId, err := identity.NewMSPIdentityFromPath(testdata.OrdererMspId, testdata.OrdererMspPath)
	if err != nil {
		panic(err)
	}
	id = mspId.GetSigningIdentity(cs)
}
