package lifecycle

import (
	"context"
	_ "embed"

	"github.com/golang/protobuf/ptypes/empty"
	lifecycleproto "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	lifecyclecc "github.com/hyperledger/fabric/core/chaincode/lifecycle"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/service"
)

//go:embed lifecycle.swagger.json
var Swagger []byte

type (
	Service struct {
		UnimplementedLifecycleServiceServer

		Invoker api.Invoker
	}
)

func NewLifecycle(invoker api.Invoker) *Service {
	return &Service{
		Invoker: invoker,
	}
}

func (l *Service) ServiceDef() *service.Def {
	return service.NewDef(
		_LifecycleService_serviceDesc.ServiceName,
		Swagger,
		&_LifecycleService_serviceDesc,
		l,
		RegisterLifecycleServiceHandlerFromEndpoint,
	)
}

func (l *Service) QueryInstalledChaincodes(ctx context.Context, _ *empty.Empty) (*lifecycleproto.QueryInstalledChaincodesResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		``, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.QueryInstalledChaincodesFuncName, []byte{}},
		&lifecycleproto.QueryInstalledChaincodesResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryInstalledChaincodesResult), nil
}

func (l *Service) QueryInstalledChaincode(ctx context.Context, args *lifecycleproto.QueryInstalledChaincodeArgs) (*lifecycleproto.QueryInstalledChaincodeResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		``, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.QueryInstalledChaincodeFuncName, args},
		&lifecycleproto.QueryInstalledChaincodeResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryInstalledChaincodeResult), nil
}

func (l *Service) InstallChaincode(ctx context.Context, args *lifecycleproto.InstallChaincodeArgs) (*lifecycleproto.InstallChaincodeResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		``, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.InstallChaincodeFuncName, args},
		&lifecycleproto.InstallChaincodeResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.InstallChaincodeResult), nil
}

func (l *Service) ApproveChaincodeDefinitionForMyOrg(ctx context.Context,
	approveChaincodeDefinitionForMyOrg *ApproveChaincodeDefinitionForMyOrgRequest) (*empty.Empty, error) {

	// approve method should be endorsed only on local msp peer
	ctxWithEndorserSpecified := tx.ContextWithEndorserMSPs(ctx,
		[]string{l.Invoker.CurrentIdentity().GetMSPIdentifier()})

	args, err := tx.ArgsBytes(
		lifecyclecc.ApproveChaincodeDefinitionForMyOrgFuncName,
		approveChaincodeDefinitionForMyOrg.Args,
	)
	if err != nil {
		return nil, err
	}

	_, _, err = l.Invoker.Invoke(
		ctxWithEndorserSpecified,
		approveChaincodeDefinitionForMyOrg.Channel,
		chaincode.Lifecycle,
		args, nil, nil, ``)
	if err != nil {
		return nil, err
	}

	return nil, err
}

func (l *Service) QueryApprovedChaincodeDefinition(ctx context.Context, queryApprovedChaincodeDefinition *QueryApprovedChaincodeDefinitionRequest) (*lifecycleproto.QueryApprovedChaincodeDefinitionResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		queryApprovedChaincodeDefinition.Channel, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.QueryApprovedChaincodeDefinitionFuncName, queryApprovedChaincodeDefinition.Args},
		&lifecycleproto.QueryApprovedChaincodeDefinitionResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryApprovedChaincodeDefinitionResult), nil
}

func (l *Service) CheckCommitReadiness(ctx context.Context, checkCommitReadiness *CheckCommitReadinessRequest) (
	*lifecycleproto.CheckCommitReadinessResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		checkCommitReadiness.Channel, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.CheckCommitReadinessFuncName, checkCommitReadiness.Args},
		&lifecycleproto.CheckCommitReadinessResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.CheckCommitReadinessResult), nil
}

func (l *Service) CommitChaincodeDefinition(ctx context.Context, commitChaincodeDefinition *CommitChaincodeDefinitionRequest) (*lifecycleproto.CommitChaincodeDefinitionResult, error) {
	res, err := tx.InvokeProto(ctx,
		l.Invoker,
		commitChaincodeDefinition.Channel, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.CommitChaincodeDefinitionFuncName, commitChaincodeDefinition.Args},
		&lifecycleproto.CommitChaincodeDefinitionResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.CommitChaincodeDefinitionResult), nil
}

func (l *Service) QueryChaincodeDefinition(ctx context.Context, queryChaincodeDefinition *QueryChaincodeDefinitionRequest) (*lifecycleproto.QueryChaincodeDefinitionResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		queryChaincodeDefinition.Channel, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.QueryChaincodeDefinitionFuncName, queryChaincodeDefinition.Args},
		&lifecycleproto.QueryChaincodeDefinitionResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryChaincodeDefinitionResult), nil
}

func (l *Service) QueryChaincodeDefinitions(ctx context.Context, queryChaincodeDefinitions *QueryChaincodeDefinitionsRequest) (*lifecycleproto.QueryChaincodeDefinitionsResult, error) {
	res, err := tx.QueryProto(ctx,
		l.Invoker,
		queryChaincodeDefinitions.Channel, chaincode.Lifecycle,
		[]interface{}{lifecyclecc.QueryChaincodeDefinitionsFuncName, queryChaincodeDefinitions.Args},
		&lifecycleproto.QueryChaincodeDefinitionsResult{})
	if err != nil {
		return nil, err
	}
	return res.(*lifecycleproto.QueryChaincodeDefinitionsResult), nil
}
