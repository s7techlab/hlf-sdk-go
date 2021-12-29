package system

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/core/chaincode/lifecycle"
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/v2/peer"
)

// NewLifecycle returns an implementation of api.Lifecycle interface
func NewLifecycle(peerPool api.PeerPool, identity msp.SigningIdentity) api.Lifecycle {
	return &lifecycleCC{peerPool: peerPool, identity: identity}
}

var _ api.Lifecycle = (*lifecycleCC)(nil)

type lifecycleCC struct {
	peerPool api.PeerPool
	identity msp.SigningIdentity
}

// QueryInstalledChaincodes returns installed chaincodes list
func (c *lifecycleCC) QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error) {
	resp, err := c.endorse(ctx, ``, lifecycle.QueryInstalledChaincodesFuncName, []byte(``))
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
	resp, err = c.endorse(ctx, ``, lifecycle.InstallChaincodeFuncName, args)
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
	channel api.Channel,
	approveArgs *lb.ApproveChaincodeDefinitionForMyOrgArgs) error {
	var (
		args []byte
		cc   api.Chaincode
		err  error
	)
	if args, err = proto.Marshal(approveArgs); err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}
	cc, err = channel.Chaincode(ctx, lifecycleName)
	if err != nil {
		return fmt.Errorf("initalize chaincode: %w", err)
	}

	_, _, err = cc.Invoke(lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName).
		WithIdentity(c.identity).
		ArgBytes([][]byte{args}).
		Do(ctx, api.WithEndorsingMpsIDs([]string{c.identity.GetMSPIdentifier()}))

	if err != nil {
		return fmt.Errorf("invoke chaincode: %w", err)
	}

	return nil
}

// CheckCommitReadiness returns commitments statuses of participants on chaincode definition
func (c *lifecycleCC) CheckCommitReadiness(ctx context.Context, channelID string, args *lb.CheckCommitReadinessArgs) (
	*lb.CheckCommitReadinessResult, error) {
	var (
		argsBytes []byte
		resp      []byte
		err       error
	)
	if argsBytes, err = proto.Marshal(args); err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	resp, err = c.endorse(ctx, channelID, lifecycle.CheckCommitReadinessFuncName, argsBytes)
	if err != nil {
		return nil, fmt.Errorf(`failed to endorse proposal: %w`, err)
	}
	result := new(lb.CheckCommitReadinessResult)
	if err = proto.Unmarshal(resp, result); err != nil {
		return nil, fmt.Errorf("unmarshal proposal response: %w", err)
	}

	return result, nil
}

// Commit the chaincode definition on the channel
func (c *lifecycleCC) Commit(ctx context.Context, channel api.Channel, commitArgs *lb.CommitChaincodeDefinitionArgs) (
	*lb.CommitChaincodeDefinitionResult, error) {
	var (
		args []byte
		cc   api.Chaincode
		resp *peer.Response
		err  error
	)
	if args, err = proto.Marshal(commitArgs); err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	cc, err = channel.Chaincode(ctx, lifecycleName)
	if err != nil {
		return nil, fmt.Errorf("initalize chaincode: %w", err)
	}

	resp, _, err = cc.Invoke(lifecycle.CommitChaincodeDefinitionFuncName).
		WithIdentity(c.identity).
		ArgBytes([][]byte{args}).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("invoke chaincode: %w", err)
	}
	result := new(lb.CommitChaincodeDefinitionResult)
	if err = proto.Unmarshal(resp.Payload, result); err != nil {
		return nil, fmt.Errorf("unmarshal peer response: %w", err)
	}

	return result, nil
}

// QueryChaincodeDefinition returns chaincode definition committed on the channel
func (c *lifecycleCC) QueryChaincodeDefinition(
	ctx context.Context, channel api.Channel, args *lb.QueryChaincodeDefinitionArgs) (
	*lb.QueryChaincodeDefinitionResult, error) {
	var (
		cc       api.Chaincode
		argBytes []byte
		err      error
		resp     *peer.Response
	)
	cc, err = channel.Chaincode(ctx, lifecycleName)
	if err != nil {
		return nil, fmt.Errorf("initalize chaincode: %w", err)
	}
	if argBytes, err = proto.Marshal(args); err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	resp, err = cc.Query(lifecycle.QueryChaincodeDefinitionFuncName).
		WithArguments([][]byte{argBytes}).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("query chancode: %w", err)
	}

	result := new(lb.QueryChaincodeDefinitionResult)
	if err = proto.Unmarshal(resp.Payload, result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}

// QueryChaincodeDefinitions returns chaincode definitions committed on the channel
func (c *lifecycleCC) QueryChaincodeDefinitions(
	ctx context.Context,
	channel api.Channel,
	args *lb.QueryChaincodeDefinitionsArgs) (
	*lb.QueryChaincodeDefinitionsResult, error) {
	var (
		cc       api.Chaincode
		argBytes []byte
		err      error
		resp     *peer.Response
	)
	cc, err = channel.Chaincode(ctx, lifecycleName)
	if err != nil {
		return nil, fmt.Errorf("initalize chaincode: %w", err)
	}
	if argBytes, err = proto.Marshal(args); err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	resp, err = cc.Query(lifecycle.QueryChaincodeDefinitionsFuncName).
		WithArguments([][]byte{argBytes}).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("query chancode: %w", err)
	}

	result := new(lb.QueryChaincodeDefinitionsResult)
	if err = proto.Unmarshal(resp.Payload, result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}

func (c *lifecycleCC) endorse(ctx context.Context, channel string, fn string, args ...[]byte) ([]byte, error) {
	processor := peerSDK.NewProcessor(channel)
	prop, _, err := processor.CreateProposal(lifecycleName, c.identity, fn, args, nil)
	if err != nil {
		return nil, fmt.Errorf(`failed to create proposal: %w`, err)
	}

	resp, err := c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, fmt.Errorf(`failed to endorse proposal: %w`, err)
	}
	return resp.Response.Payload, nil
}
