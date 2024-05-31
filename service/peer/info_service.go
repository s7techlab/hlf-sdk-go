package peer

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric-protos-go/common"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	hlf_sdk_go "github.com/s7techlab/hlf-sdk-go"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client"
	"github.com/s7techlab/hlf-sdk-go/client/channel"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/s7techlab/hlf-sdk-go/service"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
	"github.com/s7techlab/hlf-sdk-go/service/systemcc/cscc"
	qsccSvc "github.com/s7techlab/hlf-sdk-go/service/systemcc/qscc"
)

var (
	_ PeerInfoServiceServer = &InfoService{}

	ErrChaincodeNotFound = errors.New(`chaincode not founds`)
)

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

func applyInfoServiceDefaults(i *InfoService) {
	if i.timeouts == nil {
		i.timeouts = DefaultReadTimeouts
	}

	if i.logger == nil {
		i.logger = zap.NewNop()
	}

	if i.metrics == nil {
		i.metrics = NewEmptyMetricsClient()
	}
}

func (i *InfoService) ServiceDef() *service.Def {
	return service.NewDef(
		_PeerInfoService_serviceDesc.ServiceName,
		Swagger,
		&_PeerInfoService_serviceDesc,
		i,
		RegisterPeerInfoServiceHandlerFromEndpoint)
}

func (i *InfoService) peerClient(ctx context.Context) (api.Peer, error) {
	return client.NewPeer(ctx, i.connection, i.msp.Signer(), i.logger)
}

func (i *InfoService) chaincodeInfoClient(ctx context.Context, fabricVersion hlf_sdk_go.FabricVersion) (ChaincodeInfoClient, error) {
	peer, err := i.peerClient(ctx)
	if err != nil {
		return nil, err
	}

	admin := i.msp.Admins()[0]
	signer := i.msp.Signer()

	switch fabricVersion {
	case hlf_sdk_go.FabricV2:
		return &LifecycleChaincodeInfoClient{
			querier:       peer,
			admin:         admin,
			signer:        signer,
			fabricVersion: fabricVersion,
			logger:        i.logger,
		}, nil
	case hlf_sdk_go.FabricV1:
		return &LSCCChaincodeInfoClient{
			querier:       peer,
			admin:         admin,
			signer:        signer,
			fabricVersion: fabricVersion,
			logger:        i.logger,
		}, nil
	}

	return nil, hlf_sdk_go.ErrUnknownFabricVersion
}

func (i *InfoService) GetInstalledChaincodes(ctx context.Context, _ *emptypb.Empty) (*Chaincodes, error) {
	i.logger.Debug(`get installed chaincodes`)

	ctxRead, ctxReadCancel := context.WithTimeout(ctx, i.timeouts.ReadRequest)
	defer ctxReadCancel()

	ccClientLSCC, err := i.chaincodeInfoClient(ctx, hlf_sdk_go.FabricV1)
	if err != nil {
		return nil, err
	}

	chaincodes, err := ccClientLSCC.GetInstalledChaincodes(ctxRead)
	if err != nil {
		return nil, fmt.Errorf(`get lscc installed chaincodes: %w`, err)
	}

	if i.fabricVersion == hlf_sdk_go.FabricV2 {
		ccClientLifecycle, err := i.chaincodeInfoClient(ctx, hlf_sdk_go.FabricV2)
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

func (i *InfoService) GetInstalledChaincode(ctx context.Context, id *ccpackage.PackageID) (*Chaincode, error) {
	ccs, err := i.GetInstalledChaincodes(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	for _, cc := range ccs.Chaincodes {
		if cc.Name == id.Name &&
			cc.Version == id.Version && LifecycleVersionMatch(cc.LifecycleVersion, id.FabricVersion) {
			return cc, nil
		}
	}

	return nil, ErrChaincodeNotFound
}

func (i *InfoService) ListChannels(ctx context.Context, empty *emptypb.Empty) (*Channels, error) {
	peer, err := i.peerClient(ctx)
	if err != nil {
		return nil, err
	}

	channelQueryResponse, err := channel.NewCSCCListGetter(peer).GetChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("get channels: %w", err)
	}

	var (
		channels Channels
		info     *common.BlockchainInfo
	)

	qsccClient := qsccSvc.NewQSCC(peer)

	for _, ch := range channelQueryResponse.GetChannels() {
		info, err = qsccClient.GetChainInfo(ctx, &qsccSvc.GetChainInfoRequest{ChannelName: ch.ChannelId})
		if err != nil {
			return nil, fmt.Errorf("get channel=%s info using qscc: %w", ch.GetChannelId(), err)
		}
		channels.Channels = append(channels.Channels, &Channel{
			Name:   ch.GetChannelId(),
			Height: info.GetHeight(),
		})
	}

	return &channels, nil
}

func (i *InfoService) GetChannel(ctx context.Context, req *GetChannelRequest) (*GetChannelResponse, error) {
	peer, err := i.peerClient(ctx)
	if err != nil {
		return nil, err
	}

	csccClient := cscc.New(peer, i.fabricVersion)
	ctxRead, ctxReadCancel := context.WithTimeout(ctx, i.timeouts.ReadRequest)
	defer ctxReadCancel()

	config, err := csccClient.GetChannelConfig(ctxRead, &cscc.GetChannelConfigRequest{Channel: req.ChannelName})
	if err != nil {
		return nil, fmt.Errorf("fetch config block from channel=%s: %w", req.ChannelName, err)
	}

	// parse channel config
	chanDetailed, err := channel.ReadConfig(config)
	if err != nil {
		return nil, fmt.Errorf("pars channel block: %w", err)
	}
	// because name wasn't provided by channel.ReadConfig, see description
	chanDetailed.Name = getChannel.ChannelName
	chanInfo := chanDetailed.ToSimpleVersion()

	chainInfo, err := system.NewQSCC(client).GetChainInfo(ctxRead,
		&system.GetChainInfoRequest{ChannelName: getChannel.ChannelName})
	if err != nil {

		return nil, fmt.Errorf("fetch blockchain info: %w", err)
	}

	return &GetChannelResponse{
		Channel: chanInfo,
		Height:  chainInfo.Height,
	}, nil
}

func (i *InfoService) GetInstantiatedChaincodes(ctx context.Context, request *GetInstantiatedChaincodesRequest) (*Chaincodes, error) {
	//TODO implement me
	panic("implement me")
}

func (i *InfoService) GetPeerInfo(ctx context.Context, empty *emptypb.Empty) (*GetPeerInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}
