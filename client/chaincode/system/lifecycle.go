package system

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/core/chaincode/lifecycle"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/v2/peer"
)

// NewLifecycle returns an implementation of api.Lifecycle interface
func NewLifecycle(peerPool api.PeerPool, identity msp.SigningIdentity) api.Lifecycle {
	return &lifecycleCC{peerPool: peerPool, identity: identity, processor: peerSDK.NewProcessor(``)}
}

var _ api.Lifecycle = (*lifecycleCC)(nil)

type lifecycleCC struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

// QueryInstalledChaincodes returns installed chaincodes list
func (c *lifecycleCC) QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error) {
	resp, err := c.endorse(ctx, lifecycle.QueryInstalledChaincodesFuncName, []byte(``))
	if err != nil {
		return nil, err
	}
	ccData := new(lb.QueryInstalledChaincodesResult)
	if err = proto.Unmarshal(resp, ccData); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return ccData, nil
}

// InstallChaincode install chaincode on a peer
func (c *lifecycleCC) InstallChaincode(ctx context.Context, installArgs *lb.InstallChaincodeArgs) (*lb.InstallChaincodeResult, error) {
	var (
		args []byte
		resp []byte
		err  error
	)
	if args, err = proto.Marshal(installArgs); err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	resp, err = c.endorse(ctx, lifecycle.InstallChaincodeFuncName, args)
	if err != nil {
		return nil, err
	}

	ccResult := new(lb.InstallChaincodeResult)
	if err = proto.Unmarshal(resp, ccResult); err != nil {
		return nil, fmt.Errorf("unmarshal protobuf: %w", err)
	}

	return ccResult, nil
}

func (c *lifecycleCC) endorse(ctx context.Context, fn string, args ...[]byte) ([]byte, error) {
	prop, _, err := c.processor.CreateProposal(lifecycleName, c.identity, fn, args, nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
	}
	return resp.Response.Payload, nil
}
