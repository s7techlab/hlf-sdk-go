package system

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/core/chaincode/lifecycle"
	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/v2/peer"
)

var _ api.Lifecycle = (*lifecycleCoreFacade)(nil)

// About separation of implementations
// Thing is: new fabric v2 lifecycle uses discovery for some methods, and for some doesn't
// we dont want to use discovery everytime because some services must work just with one home-peer and should't(and doenst want to) know about rest topogy/configuration
func NewLifecycle(core api.Core) *lifecycleCoreFacade {
	lfn := &lifecycleNoDiscovery{
		peerPool:        core.PeerPool(),
		signingIdentity: core.CurrentIdentity(),
	}

	lfw := &lifecycleWithDiscovery{
		core: core,
	}

	return &lifecycleCoreFacade{
		lfn: lfn,
		lfw: lfw,
	}
}

type lifecycleNoDiscovery struct {
	peerPool        api.PeerPool
	signingIdentity msp.SigningIdentity
}

type lifecycleWithDiscovery struct {
	core api.Core
}

type lifecycleCoreFacade struct {
	lfn *lifecycleNoDiscovery
	lfw *lifecycleWithDiscovery
}

/* core facade */

func (l lifecycleCoreFacade) QueryInstalledChaincode(ctx context.Context, args *lb.QueryInstalledChaincodeArgs) (*lb.QueryInstalledChaincodeResult, error) {
	return l.lfn.QueryInstalledChaincode(ctx, args)
}

func (l lifecycleCoreFacade) QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error) {
	return l.lfn.QueryInstalledChaincodes(ctx)
}

func (l lifecycleCoreFacade) InstallChaincode(ctx context.Context, args *lb.InstallChaincodeArgs) (*lb.InstallChaincodeResult, error) {
	return l.lfn.InstallChaincode(ctx, args)
}

func (l lifecycleCoreFacade) ApproveChaincodeDefinitionForMyOrg(ctx context.Context, channel string, args *lb.ApproveChaincodeDefinitionForMyOrgArgs) error {
	return l.lfw.ApproveChaincodeDefinitionForMyOrg(ctx, channel, args)
}

func (l lifecycleCoreFacade) QueryApprovedChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryApprovedChaincodeDefinitionArgs) (*lb.QueryApprovedChaincodeDefinitionResult, error) {
	return l.lfn.QueryApprovedChaincodeDefinition(ctx, channel, args)
}

func (l lifecycleCoreFacade) CheckCommitReadiness(ctx context.Context, channel string, args *lb.CheckCommitReadinessArgs) (*lb.CheckCommitReadinessResult, error) {
	return l.lfn.CheckCommitReadiness(ctx, channel, args)
}

func (l lifecycleCoreFacade) CommitChaincodeDefinition(ctx context.Context, channel string, args *lb.CommitChaincodeDefinitionArgs) (*lb.CommitChaincodeDefinitionResult, error) {
	return l.lfw.CommitChaincodeDefinition(ctx, channel, args)
}

func (l lifecycleCoreFacade) QueryChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionArgs) (*lb.QueryChaincodeDefinitionResult, error) {
	return l.lfn.QueryChaincodeDefinition(ctx, channel, args)
}

func (l lifecycleCoreFacade) QueryChaincodeDefinitions(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionsArgs) (*lb.QueryChaincodeDefinitionsResult, error) {
	return l.lfn.QueryChaincodeDefinitions(ctx, channel, args)
}

/* lifecycleNoDiscovery */

func (l *lifecycleNoDiscovery) QueryInstalledChaincode(
	ctx context.Context,
	args *lb.QueryInstalledChaincodeArgs,
) (
	*lb.QueryInstalledChaincodeResult,
	error,
) {
	argBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	channel := ""

	resBytes, err := l.endorse(ctx, channel, lifecycle.QueryInstalledChaincodeFuncName, argBytes)
	if err != nil {
		return nil, err
	}

	result := &lb.QueryInstalledChaincodeResult{
		References: make(map[string]*lb.QueryInstalledChaincodeResult_References),
	}
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf(`failed to unmarshal protobuf: %w`, err)
	}

	return result, nil
}
func (l *lifecycleNoDiscovery) QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error) {
	channel := ""
	args := make([][]byte, 1)

	resBytes, err := l.endorse(ctx, channel, lifecycle.QueryInstalledChaincodesFuncName, args...)
	if err != nil {
		return nil, err
	}

	result := &lb.QueryInstalledChaincodesResult{
		InstalledChaincodes: make([]*lb.QueryInstalledChaincodesResult_InstalledChaincode, 0),
	}
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf(`failed to unmarshal protobuf: %w`, err)
	}

	return result, nil
}
func (l *lifecycleNoDiscovery) InstallChaincode(ctx context.Context, args *lb.InstallChaincodeArgs) (*lb.InstallChaincodeResult, error) {
	channel := ""
	argBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}

	resBytes, err := l.endorse(ctx, channel, lifecycle.InstallChaincodeFuncName, argBytes)
	if err != nil {
		return nil, err
	}

	result := new(lb.InstallChaincodeResult)
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf("unmarshal protobuf: %w", err)
	}

	return result, nil
}
func (l *lifecycleNoDiscovery) QueryApprovedChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryApprovedChaincodeDefinitionArgs) (*lb.QueryApprovedChaincodeDefinitionResult, error) {
	argBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}

	resBytes, err := l.endorse(ctx, channel, lifecycle.QueryApprovedChaincodeDefinitionFuncName, argBytes)
	if err != nil {
		return nil, err
	}

	result := new(lb.QueryApprovedChaincodeDefinitionResult)
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf("unmarshal proposal response: %w", err)
	}

	return result, nil
}

func (l *lifecycleNoDiscovery) CheckCommitReadiness(ctx context.Context, channel string, args *lb.CheckCommitReadinessArgs) (*lb.CheckCommitReadinessResult, error) {
	argBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}

	resBytes, err := l.endorse(ctx, channel, lifecycle.CheckCommitReadinessFuncName, argBytes)
	if err != nil {
		return nil, err
	}

	result := new(lb.CheckCommitReadinessResult)
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf("unmarshal proposal response: %w", err)
	}

	return result, nil
}

func (l *lifecycleNoDiscovery) QueryChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionArgs) (*lb.QueryChaincodeDefinitionResult, error) {
	argBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}

	resBytes, err := l.endorse(ctx, channel, lifecycle.QueryChaincodeDefinitionFuncName, argBytes)
	if err != nil {
		return nil, err
	}

	result := new(lb.QueryChaincodeDefinitionResult)
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}
func (l *lifecycleNoDiscovery) QueryChaincodeDefinitions(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionsArgs) (*lb.QueryChaincodeDefinitionsResult, error) {
	argBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}

	resBytes, err := l.endorse(ctx, channel, lifecycle.QueryChaincodeDefinitionsFuncName, argBytes)
	if err != nil {
		return nil, err
	}

	result := new(lb.QueryChaincodeDefinitionsResult)
	if err = proto.Unmarshal(resBytes, result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}

func (l *lifecycleNoDiscovery) endorse(ctx context.Context, channel string, fn string, args ...[]byte) ([]byte, error) {
	processor := peerSDK.NewProcessor(channel)
	prop, _, err := processor.CreateProposal(lifecycleName, l.signingIdentity, fn, args, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	resp, err := l.peerPool.Process(ctx, l.signingIdentity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, fmt.Errorf("failed to endorse proposal: %w", err)
	}
	return resp.Response.Payload, nil
}

/* lifecycleWithDiscovery */

func (l *lifecycleWithDiscovery) ApproveChaincodeDefinitionForMyOrg(ctx context.Context, channel string, args *lb.ApproveChaincodeDefinitionForMyOrgArgs) error {
	cc, err := l.core.Channel(channel).Chaincode(ctx, lifecycleName)
	if err != nil {
		return err
	}
	var argBytes []byte
	if argBytes, err = proto.Marshal(args); err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}

	_, _, err = cc.Invoke(lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName).
		ArgBytes([][]byte{argBytes}).
		Do(ctx, api.WithEndorsingMpsIDs([]string{l.core.CurrentIdentity().GetMSPIdentifier()}))

	if err != nil {
		return fmt.Errorf("invoke chaincode: %w", err)
	}

	return nil
}
func (l *lifecycleWithDiscovery) CommitChaincodeDefinition(ctx context.Context, channel string, args *lb.CommitChaincodeDefinitionArgs) (*lb.CommitChaincodeDefinitionResult, error) {
	cc, err := l.core.Channel(channel).Chaincode(ctx, lifecycleName)
	if err != nil {
		return nil, err
	}

	argsBytes, err := proto.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}

	resp, _, err := cc.Invoke(lifecycle.CommitChaincodeDefinitionFuncName).
		ArgBytes([][]byte{argsBytes}).
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
