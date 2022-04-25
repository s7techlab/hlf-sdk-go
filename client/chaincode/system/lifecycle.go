package system

import (
	"context"
	_ "embed"

	"github.com/golang/protobuf/ptypes/empty"
	lifecycleproto "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	lifecyclecc "github.com/hyperledger/fabric/core/chaincode/lifecycle"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
)

//go:embed lifecycle.swagger.json
var LifecycleServiceSwagger []byte

type (
	LifecycleService struct {
		UnimplementedLifecycleServiceServer

		Invoker api.Invoker
	}
)

func NewLifecycle(invoker api.Invoker) *LifecycleService {
	return &LifecycleService{
		Invoker: invoker,
	}
}

func (l *LifecycleService) ServiceDef() ServiceDef {
	return NewServiceDef(
		_LifecycleService_serviceDesc.ServiceName,
		LifecycleServiceSwagger,
		&_LifecycleService_serviceDesc,
		l,
		RegisterLifecycleServiceHandlerFromEndpoint,
	)
}

func (l *LifecycleService) QueryInstalledChaincodes(ctx context.Context, _ *empty.Empty) (*lifecycleproto.QueryInstalledChaincodesResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		``, LifecycleName,
		[]interface{}{lifecyclecc.QueryInstalledChaincodesFuncName, []byte{}},
		&lifecycleproto.QueryInstalledChaincodesResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryInstalledChaincodesResult), nil
}

func (l *LifecycleService) QueryInstalledChaincode(ctx context.Context, args *lifecycleproto.QueryInstalledChaincodeArgs) (*lifecycleproto.QueryInstalledChaincodeResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		``, LifecycleName,
		[]interface{}{lifecyclecc.QueryInstalledChaincodeFuncName, args},
		&lifecycleproto.QueryInstalledChaincodeResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryInstalledChaincodeResult), nil
}

func (l *LifecycleService) InstallChaincode(ctx context.Context, args *lifecycleproto.InstallChaincodeArgs) (*lifecycleproto.InstallChaincodeResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		``, LifecycleName,
		[]interface{}{lifecyclecc.InstallChaincodeFuncName, args},
		&lifecycleproto.InstallChaincodeResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.InstallChaincodeResult), nil
}

func (l *LifecycleService) ApproveChaincodeDefinitionForMyOrg(ctx context.Context, approveChaincodeDefinitionForMyOrg *ApproveChaincodeDefinitionForMyOrgRequest) (*empty.Empty, error) {

	// for invoker need to set endorser msp
	// Do(ctx, api.WithEndorsingMpsIDs([]string{l.core.CurrentIdentity().GetMSPIdentifier()}))
	args, err := tx.ArgsBytes(lifecyclecc.ApproveChaincodeDefinitionForMyOrgFuncName, approveChaincodeDefinitionForMyOrg.Args)
	if err != nil {
		return nil, err
	}

	_, _, err = l.Invoker.Invoke(ctx, approveChaincodeDefinitionForMyOrg.Channel, LifecycleName, args, nil, nil, ``)
	if err != nil {
		return nil, err
	}

	return nil, err
}

func (l *LifecycleService) QueryApprovedChaincodeDefinition(ctx context.Context, queryApprovedChaincodeDefinition *QueryApprovedChaincodeDefinitionRequest) (*lifecycleproto.QueryApprovedChaincodeDefinitionResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		queryApprovedChaincodeDefinition.Channel, LifecycleName,
		[]interface{}{lifecyclecc.QueryApprovedChaincodeDefinitionFuncName, queryApprovedChaincodeDefinition.Args},
		&lifecycleproto.QueryApprovedChaincodeDefinitionResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryApprovedChaincodeDefinitionResult), nil
}

func (l *LifecycleService) CheckCommitReadiness(ctx context.Context, сheckCommitReadiness *CheckCommitReadinessRequest) (*lifecycleproto.CheckCommitReadinessResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		сheckCommitReadiness.Channel, LifecycleName,
		[]interface{}{lifecyclecc.CheckCommitReadinessFuncName, сheckCommitReadiness.Args},
		&lifecycleproto.CheckCommitReadinessResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.CheckCommitReadinessResult), nil
}

func (l *LifecycleService) CommitChaincodeDefinition(ctx context.Context, commitChaincodeDefinition *CommitChaincodeDefinitionRequest) (*lifecycleproto.CommitChaincodeDefinitionResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		commitChaincodeDefinition.Channel, LifecycleName,
		[]interface{}{lifecyclecc.CommitChaincodeDefinitionFuncName, commitChaincodeDefinition.Args},
		&lifecycleproto.CommitChaincodeDefinitionResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.CommitChaincodeDefinitionResult), nil
}

func (l *LifecycleService) QueryChaincodeDefinition(ctx context.Context, queryChaincodeDefinition *QueryChaincodeDefinitionRequest) (*lifecycleproto.QueryChaincodeDefinitionResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		queryChaincodeDefinition.Channel, LifecycleName,
		[]interface{}{lifecyclecc.QueryChaincodeDefinitionFuncName, queryChaincodeDefinition.Args},
		&lifecycleproto.QueryChaincodeDefinitionResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryChaincodeDefinitionResult), nil
}

func (l *LifecycleService) QueryChaincodeDefinitions(ctx context.Context, queryChaincodeDefinitions *QueryChaincodeDefinitionsRequest) (*lifecycleproto.QueryChaincodeDefinitionsResult, error) {

	res, err := tx.QueryProto(ctx,
		l.Invoker,
		queryChaincodeDefinitions.Channel, LifecycleName,
		[]interface{}{lifecyclecc.QueryChaincodeDefinitionsFuncName, queryChaincodeDefinitions.Args},
		&lifecycleproto.QueryChaincodeDefinitionsResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryChaincodeDefinitionsResult), nil
}
