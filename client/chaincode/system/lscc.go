package system

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric-protos-go/peer"
	lsccPkg "github.com/hyperledger/fabric/core/scc/lscc"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/client/tx"
)

//go:embed lscc.swagger.json
var LSCCServiceSwagger []byte

type (
	LSCCService struct {
		UnimplementedLSCCServiceServer

		Invoker api.Invoker
	}
)

func NewLSCC(invoker api.Invoker) *LSCCService {
	return &LSCCService{
		Invoker: invoker,
	}
}

func (l *LSCCService) ServiceDef() ServiceDef {
	return NewServiceDef(
		_LSCCService_serviceDesc.ServiceName,
		LSCCServiceSwagger,
		&_LSCCService_serviceDesc,
		l,
		RegisterLSCCServiceHandlerFromEndpoint,
	)
}

func (l *LSCCService) GetChaincodeData(ctx context.Context, getChaincodeData *GetChaincodeDataRequest) (*peer.ChaincodeData, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		getChaincodeData.Channel, LSCCName,
		[]string{lsccPkg.GETCCDATA, getChaincodeData.Channel, getChaincodeData.Chaincode},
		&peer.ChaincodeData{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeData), nil
}

func (l *LSCCService) GetInstalledChaincodes(ctx context.Context, _ *empty.Empty) (*peer.ChaincodeQueryResponse, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		``, LSCCName,
		[]string{lsccPkg.GETINSTALLEDCHAINCODES},
		&peer.ChaincodeQueryResponse{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeQueryResponse), nil
}
func (l *LSCCService) GetChaincodes(ctx context.Context, getChaincodes *GetChaincodesRequest) (*peer.ChaincodeQueryResponse, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		getChaincodes.Channel, LSCCName,
		[]string{lsccPkg.GETCHAINCODES},
		&peer.ChaincodeQueryResponse{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeQueryResponse), nil
}

func (l *LSCCService) GetDeploymentSpec(ctx context.Context, getDeploymentSpec *GetDeploymentSpecRequest) (*peer.ChaincodeDeploymentSpec, error) {
	res, err := tx.QueryStringsProto(ctx,
		l.Invoker,
		getDeploymentSpec.Channel, LSCCName,
		[]string{lsccPkg.GETDEPSPEC, getDeploymentSpec.Channel, getDeploymentSpec.Chaincode},
		&peer.ChaincodeDeploymentSpec{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChaincodeDeploymentSpec), nil
}
func (l *LSCCService) Install(ctx context.Context, spec *peer.ChaincodeDeploymentSpec) (*empty.Empty, error) {
	_, err := tx.QueryProto(ctx,
		l.Invoker,
		``, LSCCName,
		[]interface{}{lsccPkg.INSTALL, spec},
		&peer.ChaincodeDeploymentSpec{})

	return nil, err
}
func (l *LSCCService) Deploy(ctx context.Context, deploy *DeployRequest) (response *peer.Response, err error) {

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
	res, _, err := l.Invoker.Invoke(ctx, deploy.Channel, LSCCName, argsBytes, nil, deploy.Transient, ``)
	return res, err
}
