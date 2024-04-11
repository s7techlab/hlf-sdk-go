package peer

import (
	"context"

	"go.uber.org/zap"

	hlf_sdk_go "github.com/s7techlab/hlf-sdk-go"
	"github.com/s7techlab/hlf-sdk-go/api/config"
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

func (p *InfoService) ccClient(ctx context.Context, fabricVersion hlf_sdk_go.FabricVersion) (ChaincodeInfoClient, error) {

	client, err := p.Client(ctx)
	if err != nil {
		return nil, err
	}

	admin := p.msp.Admins()[0]
	signer := p.msp.Signer()

	switch fabricVersion {
	case  ccpackage.FabricVersion_FABRIC_V2_LIFECYCLE
		return &LifecycleClient{
			invoker:       client,
			admin:         admin,
			signer:        signer,
			fabricVersion: fabricVersion,
			logger:        p.logger,
		}, nil
	case chaincode.FabricVersion_FABRIC_V1:
		fallthrough
	case chaincode.FabricVersion_FABRIC_V2:
		return &LSCCClient{
			invoker:       client,
			admin:         admin,
			signer:        signer,
			fabricVersion: fabricVersion,
			logger:        p.logger,
		}, nil
	}

	return nil, errors.New(`unknown fabric version`)
}
