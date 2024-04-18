package peer

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	hlf_sdk_go "github.com/s7techlab/hlf-sdk-go"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/s7techlab/hlf-sdk-go/service"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
)

var _ PeerInfoServiceServer = &InfoService{}

type Opt func(*InfoService)

func WithReadTimeouts(timeouts *ReadTimeouts) Opt {
	return func(p *InfoService) {
		p.timeouts = timeouts
	}
}

func WithLogger(logger *zap.Logger) Opt {
	return func(p *InfoService) {
		p.logger = logger
	}
}

type InfoService struct {
	msp        identity.MSP
	connection config.ConnectionConfig

	fabricVersion hlf_sdk_go.FabricVersion
	metrics       Metrics
	timeouts      *ReadTimeouts
	logger        *zap.Logger
}

func NewInfoService(msp identity.MSP, connection config.ConnectionConfig, opts ...Opt) *InfoService {
	p := &InfoService{
		msp:        msp,
		connection: connection,
	}
	for _, opt := range opts {
		opt(p)
	}
	applyInfoServiceDefaults(p)
	return p
}

func applyInfoServiceDefaults(p *InfoService) {
	if p.timeouts == nil {
		p.timeouts = DefaultReadTimeouts
	}

	if p.logger == nil {
		p.logger = zap.NewNop()
	}

	if p.metrics == nil {
		p.metrics = NewMockMetricsClient()
	}
}

func (p *InfoService) ServiceDef() *service.Def {
	return service.NewDef(
		_PeerInfoService_serviceDesc.ServiceName,
		Swagger,
		&_PeerInfoService_serviceDesc,
		p,
		RegisterPeerInfoServiceHandlerFromEndpoint)
}

func (p *InfoService) chaincodeInfoClient(ctx context.Context, fabricVersion hlf_sdk_go.FabricVersion) (ChaincodeInfoClient, error) {
	peer, err := client.NewPeer(ctx, p.connection, p.msp.Signer(), p.logger)
	if err != nil {
		return nil, err
	}

	admin := p.msp.Admins()[0]
	signer := p.msp.Signer()

	switch fabricVersion {
	case hlf_sdk_go.FabricV2:
		return &LifecycleChaincodeInfoClient{
			querier:       peer,
			admin:         admin,
			signer:        signer,
			fabricVersion: fabricVersion,
			logger:        p.logger,
		}, nil
	case hlf_sdk_go.FabricV1:
		return &LSCCChaincodeInfoClient{
			querier:       peer,
			admin:         admin,
			signer:        signer,
			fabricVersion: fabricVersion,
			logger:        p.logger,
		}, nil
	}

	return nil, hlf_sdk_go.ErrUnknownFabricVersion
}

func (p *InfoService) GetInstalledChaincodes(ctx context.Context, _ *emptypb.Empty) (*Chaincodes, error) {
	p.logger.Debug(`get installed chaincodes`)

	ctxRead, ctxReadCancel := context.WithTimeout(ctx, p.timeouts.ReadRequest)
	defer ctxReadCancel()

	ccClientLSCC, err := p.chaincodeInfoClient(ctx, hlf_sdk_go.FabricV1)
	if err != nil {
		return nil, err
	}

	chaincodes, err := ccClientLSCC.GetInstalledChaincodes(ctxRead)
	if err != nil {
		return nil, fmt.Errorf(`get lscc installed chaincodes: %w`, err)
	}

	if p.fabricVersion == hlf_sdk_go.FabricV2 {
		ccClientLifecycle, err := p.chaincodeInfoClient(ctx, hlf_sdk_go.FabricV2)
		if err != nil {
			return nil, err
		}

		lifecycleChaincodes, err := ccClientLifecycle.GetInstalledChaincodes(ctxRead)
		if err != nil {
			return nil, fmt.Errorf(`get lifecycle installed chaincodes: %w`, err)
		}
		chaincodes.Chaincodes = append(chaincodes.Chaincodes, lifecycleChaincodes.Chaincodes...)
	}

	return chaincodes, nil
}

func (p *InfoService) GetInstalledChaincode(ctx context.Context, id *ccpackage.PackageID) (*Chaincode, error) {
	//TODO implement me
	panic("implement me")
}

func (p *InfoService) ListChannels(ctx context.Context, empty *emptypb.Empty) (*Channels, error) {
	//TODO implement me
	panic("implement me")
}

func (p *InfoService) GetChannel(ctx context.Context, request *GetChannelRequest) (*GetChannelResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *InfoService) GetInstantiatedChaincodes(ctx context.Context, request *GetInstantiatedChaincodesRequest) (*Chaincodes, error) {
	//TODO implement me
	panic("implement me")
}

func (p *InfoService) GetPeerInfo(ctx context.Context, empty *emptypb.Empty) (*GetPeerInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}
