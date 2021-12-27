package system

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/core/chaincode/lifecycle"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/client/chaincode/txwaiter"
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
		return nil, fmt.Errorf(`failed to unmarshal protobuf: %w`, err)
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

// ApproveFromMyOrg approves chaincode package on a channel
func (c *lifecycleCC) ApproveFromMyOrg(
	ctx context.Context,
	channelID string,
	broadcastClient api.Orderer,
	approveArgs *lb.ApproveChaincodeDefinitionForMyOrgArgs,
) error {
	var (
		args      []byte
		resp      *peer.ProposalResponse
		processor api.PeerProcessor
		tx        api.ChaincodeTx
		err       error
	)

	if args, err = proto.Marshal(approveArgs); err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}
	processor = peerSDK.NewProcessor(channelID)

	prop, tx, err := processor.CreateProposal(
		lifecycleName,
		c.identity,
		lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName,
		[][]byte{args},
		nil,
	)
	if err != nil {
		return fmt.Errorf(`failed to create proposal: %w`, err)
	}

	resp, err = c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return fmt.Errorf(`failed to endorse proposal: %w`, err)
	}

	peerProp := new(peer.Proposal)
	err = proto.Unmarshal(prop.ProposalBytes, peerProp)
	if err != nil {
		return fmt.Errorf(`failed to unmarshal proposal: %w`, err)
	}

	env, err := protoutil.CreateSignedTx(peerProp, c.identity, resp)
	if err != nil {
		return fmt.Errorf(`create signed transaction: %w`, err)
	}
	waiter := txwaiter.NewSelfPeerWaiter(c.peerPool, c.identity)

	if _, err = broadcastClient.Broadcast(ctx, env); err != nil {
		return fmt.Errorf("broadcast envelope: %w", err)
	}

	err = waiter.Wait(ctx, channelID, tx)
	if err != nil {
		return fmt.Errorf("waiting for transaction: %w", err)
	}

	return nil
}

func (c *lifecycleCC) endorse(ctx context.Context, fn string, args ...[]byte) ([]byte, error) {
	prop, _, err := c.processor.CreateProposal(lifecycleName, c.identity, fn, args, nil)
	if err != nil {
		return nil, fmt.Errorf(`failed to create proposal: %w`, err)
	}

	resp, err := c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, fmt.Errorf(`failed to endorse proposal: %w`, err)
	}
	return resp.Response.Payload, nil
}
