// +build !fabric2

package system

import (
	"context"
	lifecycle3 "github.com/hyperledger/fabric-protos-go/peer/lifecycle"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/util"
	lifecycle2 "github.com/hyperledger/fabric/core/chaincode/lifecycle"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

type lifecycle struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

func (c *lifecycle) QueryInstalledChaincodes(ctx context.Context) (*lifecycle3.QueryInstalledChaincodesResult, error) {
	resp, err := c.endorse(ctx, lifecycle2.QueryInstalledChaincodesFuncName, ``)
	if err != nil {
		return nil, err
	}
	ccData := new(lifecycle3.QueryInstalledChaincodesResult)
	if err = proto.Unmarshal(resp, ccData); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return ccData, nil
}

func (c *lifecycle) endorse(ctx context.Context, fn string, args ...string) ([]byte, error) {
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
	return &lifecycle{peerPool: peerPool, identity: identity, processor: peerSDK.NewProcessor(``)}
}
