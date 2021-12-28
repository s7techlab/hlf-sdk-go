package system

import (
	"context"

	"github.com/golang/protobuf/proto"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/lifecycle"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

type lifecycleCC struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

func (c *lifecycleCC) QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error) {
	resp, err := c.endorse(ctx, lifecycle.QueryInstalledChaincodesFuncName, ``)
	if err != nil {
		return nil, err
	}
	ccData := new(lb.QueryInstalledChaincodesResult)
	if err = proto.Unmarshal(resp, ccData); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return ccData, nil
}

func (c *lifecycleCC) endorse(ctx context.Context, fn string, args ...string) ([]byte, error) {
	prop, _, err := c.processor.CreateProposal(&api.DiscoveryChaincode{Name: lifecycleName, Type: api.CCTypeGoLang}, c.identity, fn, util.ToChaincodeArgs(args...), nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
	}
	return resp.Response.Payload, nil
}

func NewLifecycle(peerPool api.PeerPool, identity msp.SigningIdentity) api.Lifecycle {
	return &lifecycleCC{peerPool: peerPool, identity: identity, processor: peerSDK.NewProcessor(``)}
}
