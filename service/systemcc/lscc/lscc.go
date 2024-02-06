package lscc

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric-protos-go/peer"
	lsccPkg "github.com/hyperledger/fabric/core/scc/lscc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/service"
)

//go:embed lscc.swagger.json
var Swagger []byte

type (
	Service struct {
		UnimplementedLSCCServiceServer

		Invoker api.Invoker
	}
)

func New(invoker api.Invoker) *Service {
	return &Service{
		Invoker: invoker,
	}
}

func (l *Service) ServiceDef() *service.Def {
	return service.NewDef(
		_LSCCService_serviceDesc.ServiceName,
		Swagger,
		&_LSCCService_serviceDesc,
		l,
		RegisterLSCCServiceHandlerFromEndpoint,
	)
}

func (l *Service) GetChaincodeData(ctx context.Context, getChaincodeData *GetChaincodeDataRequest) (*peer.ChaincodeData, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		getChaincodeData.Channel, chaincode.LSCC,
		[]string{lsccPkg.GETCCDATA, getChaincodeData.Channel, getChaincodeData.Chaincode},
		&peer.ChaincodeData{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeData), nil
}

func (l *Service) GetInstalledChaincodes(ctx context.Context, _ *empty.Empty) (*peer.ChaincodeQueryResponse, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		``, chaincode.LSCC,
		[]string{lsccPkg.GETINSTALLEDCHAINCODES},
		&peer.ChaincodeQueryResponse{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeQueryResponse), nil
}
func (l *Service) GetChaincodes(ctx context.Context, getChaincodes *GetChaincodesRequest) (*peer.ChaincodeQueryResponse, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		getChaincodes.Channel, chaincode.LSCC,
		[]string{lsccPkg.GETCHAINCODES},
		&peer.ChaincodeQueryResponse{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeQueryResponse), nil
}

func (l *Service) GetDeploymentSpec(ctx context.Context, getDeploymentSpec *GetDeploymentSpecRequest) (*peer.ChaincodeDeploymentSpec, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		getDeploymentSpec.Channel, chaincode.LSCC,
		[]string{lsccPkg.GETDEPSPEC, getDeploymentSpec.Channel, getDeploymentSpec.Chaincode},
		&peer.ChaincodeDeploymentSpec{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeDeploymentSpec), nil
}
func (l *Service) Install(ctx context.Context, spec *peer.ChaincodeDeploymentSpec) (*empty.Empty, error) {
	_, err := tx.QueryProto(ctx,
		l.Invoker,
		``, chaincode.LSCC,
		[]interface{}{lsccPkg.INSTALL, spec},
		&peer.ChaincodeDeploymentSpec{})

	return nil, err
}
func (l *Service) Deploy(ctx context.Context, deploy *DeployRequest) (response *peer.Response, err error) {

	// Find chaincode instantiated or not
	ccList, err := l.GetChaincodes(ctx, &GetChaincodesRequest{Channel: deploy.Channel})
	if err != nil {
		return nil, fmt.Errorf(`get chaincodes: %w`, err)
	}
	lsccCmd := lsccPkg.DEPLOY

	for _, cc := range ccList.Chaincodes {
		if cc.Name == deploy.DeploymentSpec.ChaincodeSpec.ChaincodeId.Name {
			lsccCmd = lsccPkg.UPGRADE
			break
		}
	}

	args := []interface{}{lsccCmd, deploy.Channel, deploy.DeploymentSpec, deploy.Policy}

	if deploy.ESCC != `` {
		args = append(args, deploy.ESCC)
	}

	if deploy.VSCC != `` {
		args = append(args, deploy.VSCC)
	}

	if deploy.CollectionConfig != nil {
		args = append(args, deploy.CollectionConfig)
	}

	argsBytes, err := tx.ArgsBytes(args...)
	if err != nil {
		return nil, fmt.Errorf(`args: %w`, err)
	}
	// Invoke here (with broadcast to orderer)
	res, _, err := l.Invoker.Invoke(ctx, deploy.Channel, chaincode.LSCC, argsBytes, nil, deploy.Transient, ``)
	return res, err
}
