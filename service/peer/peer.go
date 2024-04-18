package peer

import (
	_ "embed"
	"errors"
)

// import (
//
//	"context"

//	"errors"
//	"fmt"
//	"net"
//	"sync"
//	"time"
//
//	"github.com/google/certificate-transparency-go/x509util"
//	"github.com/hyperledger/fabric-protos-go/common"
//	"github.com/s7techlab/cckit/gateway"
//	sdkconfig "github.com/s7techlab/hlf-sdk-go/api/config"
//	sdkclient "github.com/s7techlab/hlf-sdk-go/client"
//	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
//	"github.com/s7techlab/hlf-sdk-go/identity"
//	"go.uber.org/zap"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
//	empty "google.golang.org/protobuf/types/known/emptypb"
//
//	hlfconfig "b2bchain.tech/pkg/hlf/config"
//	pkgnet "b2bchain.tech/pkg/net"
//	"b2bchain.tech/platform/network/pkg/channel"
//	peerPkg "b2bchain.tech/platform/network/pkg/peer"
//	"b2bchain.tech/platform/network/service/chaincode"
//
// )
//

//go:embed peer.swagger.json

var Swagger []byte
var (
	ErrChaincodeNotInstalled     = errors.New(`chaincode not installed`)
	ErrChaincodeAlreadyInstalled = errors.New(`chaincode already installed`)
	ErrPackageIDOrSpecRequired   = errors.New(`package id or spec required`)
)

//type opts struct {
//	timeouts      Timeouts
//	logger        *zap.Logger
//	packageServer ccpackage.PackageServiceServer
//}

//
//func (p *PeerService) InstallChaincode(ctx context.Context, installChaincode *InstallChaincodeRequest) (*Chaincode, error) {
//	if err := installChaincode.Validate(); err != nil {
//		return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
//
//	packageID := installChaincode.GetChaincodePackageId()
//	packageSpec := installChaincode.GetChaincodePackageSpec()
//
//	if packageID == nil && packageSpec == nil {
//		return nil, ErrPackageIDOrSpecRequired
//	}
//
//	p.logger.Info(`install chaincode`,
//		zap.Reflect(`package id`, packageID), zap.Reflect(`package spec`, packageSpec))
//
//	if packageSpec != nil && packageID == nil {
//		packageID = packageSpec.Id
//	}
//
//	p.logger.Debug(`request installed chaincode`, zap.Reflect(`package id`, packageID))
//	_, err := p.GetInstalledChaincode(ctx, packageID)
//	if err == nil {
//		return nil, fmt.Errorf(
//			`chaincode=%s, version=%s, fabric version=%s: %w`,
//			packageID.Name, packageID.Version, packageID.FabricVersion, ErrChaincodeAlreadyInstalled)
//	}
//	if err != nil && !errors.Is(err, ErrChaincodeNotInstalled) {
//		return nil, fmt.Errorf(`get installed chaincode=%s, version=: %w`, packageID.Name, err)
//	}
//
//	if packageSpec != nil {
//		ctxCreatePackage, ctxCreatePackageCancel := context.WithTimeout(ctx, p.timeouts.ChaincodeCreatePackage)
//		defer ctxCreatePackageCancel()
//
//		_, pkgErr := p.packages.GetOrCreate(ctxCreatePackage, packageSpec)
//		if pkgErr != nil {
//			return nil, fmt.Errorf("get or create package name=%s, version=%s, fabric_version=%s: %w",
//				packageID.Name, packageID.Version, packageID.FabricVersion, pkgErr)
//		}
//	}
//
//	ctxRead, ctxReadCancel := context.WithTimeout(ctx, p.timeouts.ReadRequest)
//	defer ctxReadCancel()
//
//	p.logger.Debug(`request deployment spec`, zap.Reflect(`package id`, packageID))
//	deploymentSpec, err := p.packages.GetDeploymentSpec(ctxRead, packageID)
//	if err != nil {
//		return nil, fmt.Errorf("get deployment spec for package name=%s, version=%s, fabric_version=%s: %w",
//			packageID.Name, packageID.Version, packageID.FabricVersion, err)
//	}
//
//	ccClient, err := p.ccClient(ctx, packageID.FabricVersion)
//	if err != nil {
//		return nil, err
//	}
//
//	ctxInstall, ctxInstallCancel := context.WithTimeout(ctx, p.timeouts.ChaincodeInstall)
//	defer ctxInstallCancel()
//
//	if err = ccClient.InstallChaincode(ctxInstall, deploymentSpec); err != nil {
//		return nil, err
//	}
//
//	return p.GetInstalledChaincode(ctx, packageID)
//}
//
//func (p *PeerService) UpChaincode(ctx context.Context, upChaincode *UpChaincodeRequest) (*UpChaincodeResponse, error) {
//	if err := upChaincode.Validate(); err != nil {
//		return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
//
//	packageID := upChaincode.GetChaincodePackageId()
//	packageSpec := upChaincode.GetChaincodePackageSpec()
//
//	if packageID == nil && packageSpec == nil {
//		return nil, ErrPackageIDOrSpecRequired
//	}
//
//	p.logger.Info(`up chaincode`,
//		zap.String(`channel`, upChaincode.Channel),
//		zap.Reflect(`package id`, packageID),
//		zap.Reflect(`package spec`, packageSpec))
//
//	if packageSpec != nil && packageID == nil {
//		packageID = packageSpec.Id
//	}
//
//	cc, err := p.GetInstalledChaincode(ctx, packageID)
//	if err != nil {
//
//		if !errors.Is(err, ErrChaincodeNotInstalled) {
//			return nil, err
//		}
//
//		installRequest := &InstallChaincodeRequest{}
//		if upChaincode.GetChaincodePackageSpec() != nil {
//			installRequest.ChaincodePackage = &InstallChaincodeRequest_ChaincodePackageSpec{
//				ChaincodePackageSpec: upChaincode.GetChaincodePackageSpec()}
//		} else {
//			installRequest.ChaincodePackage = &InstallChaincodeRequest_ChaincodePackageId{
//				ChaincodePackageId: upChaincode.GetChaincodePackageId()}
//		}
//
//		cc, err = p.InstallChaincode(ctx, installRequest)
//		if err != nil {
//			return nil, fmt.Errorf(`install chaincode: %w`, err)
//		}
//	}
//
//	ctxUpChaincode, ctxUpCancel := context.WithTimeout(ctx, p.timeouts.ChaincodeUp)
//	defer ctxUpCancel()
//
//	// here method signatures for "up" chaincode func really differs between LSCC and Lifecycle
//	p.logger.Debug(`chaincode installed, up`, zap.Reflect(`chaincode`, cc))
//
//	ccClient, err := p.ccClient(ctx, packageID.FabricVersion)
//	if err != nil {
//		return nil, err
//	}
//
//	switch packageID.FabricVersion {
//
//	case chaincode.FabricVersion_FABRIC_V2_LIFECYCLE:
//		return ccClient.(LifecycleChaincodeUpper).UpChaincode(ctxUpChaincode, cc, upChaincode)
//	case chaincode.FabricVersion_FABRIC_V1:
//		fallthrough
//	case chaincode.FabricVersion_FABRIC_V2:
//
//		deploymentSpec, err := p.packages.GetDeploymentSpec(ctxUpChaincode, packageID)
//		if err != nil {
//			return nil, fmt.Errorf(`get deployment spec: %w`, err)
//		}
//
//		return ccClient.(LSCCChaincodeUpper).
//			UpChaincode(ctxUpChaincode, deploymentSpec, upChaincode)
//	default:
//		return nil, errors.New(`fabric version not supported`)
//	}
//}
//
//func (p *PeerService) GetInstalledChaincodes(ctx context.Context, _ *empty.Empty) (*Chaincodes, error) {

//}
//
//func (p *PeerService) GetInstalledChaincode(ctx context.Context, packageID *chaincode.PackageID) (*Chaincode, error) {
//	ccs, err := p.GetInstalledChaincodes(ctx, &empty.Empty{})
//	if err != nil {
//		return nil, err
//	}
//
//	for _, cc := range ccs.Chaincodes {
//		if cc.Name == packageID.Name &&
//			cc.Version == packageID.Version &&
//			LifecycleVersionMatch(cc.LifecycleVersion, packageID.FabricVersion) {
//			return cc, nil
//		}
//	}
//
//	return nil, fmt.Errorf(`chaincode name=%s, version=%s: %w`, packageID.Name, packageID.Version, ErrChaincodeNotInstalled)
//}
//
//func (p *PeerService) GetInstantiatedChaincodes(ctx context.Context, getChaincodes *GetInstantiatedChaincodesRequest) (*Chaincodes, error) {
//	if err := getChaincodes.Validate(); err != nil {
//		return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
//
//	ccClientLSCC, err := p.ccClient(ctx, chaincode.FabricVersion_FABRIC_V1)
//	if err != nil {
//		return nil, err
//	}
//
//	ctxRead, ctxReadCancel := context.WithTimeout(ctx, p.timeouts.ReadRequest)
//	defer ctxReadCancel()
//
//	chaincodes, err := ccClientLSCC.GetInstantiatedChaincodes(ctxRead, getChaincodes.ChannelName)
//	if err != nil {
//		return nil, err
//	}
//
//	if p.fabricVersion == hlfconfig.FabricV2 {
//
//		ccClientLifecycle, err := p.ccClient(ctx, chaincode.FabricVersion_FABRIC_V2_LIFECYCLE)
//		if err != nil {
//			return nil, err
//		}
//
//		chaincodesLifecycle, err := ccClientLifecycle.GetInstantiatedChaincodes(ctxRead, getChaincodes.ChannelName)
//		if err != nil {
//			return nil, err
//		}
//		chaincodes.Chaincodes = append(chaincodes.Chaincodes, chaincodesLifecycle.Chaincodes...)
//	}
//
//	return chaincodes, nil
//}
//
//func (p *PeerService) JoinChannel(ctx context.Context, joinChannel *JoinChannelRequest) (*empty.Empty, error) {
//	if err := joinChannel.Validate(); err != nil {
//		return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
//	_, _, err := net.SplitHostPort(joinChannel.OrdererAddress)
//	if err != nil {
//		return nil, status.Error(codes.InvalidArgument, "orderer address is invalid")
//	}
//
//	client, err := p.Client(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	channels, err := system.NewCSCCFromClient(client).GetChannels(ctx, &empty.Empty{})
//	if err != nil {
//		return nil, fmt.Errorf("get channels: %w", err)
//	}
//
//	for _, channelInfo := range channels.Channels {
//		if channelInfo.ChannelId == joinChannel.ChannelId {
//			return nil, status.Error(codes.AlreadyExists, "channel has already been joined to peer")
//		}
//	}
//
//	ord, err := sdkclient.NewOrderer(ctx, sdkconfig.ConnectionConfig{
//		Host:    joinChannel.OrdererAddress,
//		Timeout: sdkconfig.Duration{Duration: p.timeouts.ChannelJoin}},
//		p.logger)
//	if err != nil {
//		return nil, fmt.Errorf("create orderer: %w", err)
//	}
//
//	apiChannel := sdkclient.NewChannel(
//		client.CurrentIdentity().GetMSPIdentifier(),
//		joinChannel.ChannelId,
//		client.PeerPool(), ord, nil,
//		client.CurrentIdentity(),
//		false,
//		p.logger)
//
//	if err = apiChannel.Join(ctx); err != nil {
//		return nil, fmt.Errorf("join channel=%s: %w", joinChannel.ChannelId, err)
//	}
//
//	return &empty.Empty{}, nil
//}
//
//// GetChannel information from peer
//func (p *PeerService) GetChannel(ctx context.Context, getChannel *GetChannelRequest) (*GetChannelResponse, error) {
//	if err := getChannel.Validate(); err != nil {
//		return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
//
//	client, err := p.Client(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	// get channel config
//	ctxRead, ctxReadCancel := context.WithTimeout(ctx, p.timeouts.ReadRequest)
//	defer ctxReadCancel()
//
//	config, err := system.NewCSCCFromClient(client).GetChannelConfig(ctxRead, &system.GetChannelConfigRequest{Channel: getChannel.ChannelName})
//	if err != nil {
//		return nil, fmt.Errorf("fetch config block: %w", err)
//	}
//
//	// parse channel config
//	chanDetailed, err := channel.ReadConfig(config)
//	if err != nil {
//		return nil, fmt.Errorf("pars channel block: %w", err)
//	}
//	// because name wasn't provided by channel.ReadConfig, see description
//	chanDetailed.Name = getChannel.ChannelName
//	chanInfo := chanDetailed.ToSimpleVersion()
//
//	chainInfo, err := system.NewQSCC(client).GetChainInfo(ctxRead,
//		&system.GetChainInfoRequest{ChannelName: getChannel.ChannelName})
//	if err != nil {
//
//		return nil, fmt.Errorf("fetch blockchain info: %w", err)
//	}
//
//	return &GetChannelResponse{
//		Channel: chanInfo,
//		Height:  chainInfo.Height,
//	}, nil
//}
//
//func (p *PeerService) ListChannels(ctx context.Context, _ *empty.Empty) (*Channels, error) {
//	client, err := p.Client(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	res, err := system.NewCSCCFromClient(client).GetChannels(ctx, &empty.Empty{})
//	if err != nil {
//		return nil, fmt.Errorf("get channels: %w", err)
//	}
//
//	var (
//		channels Channels
//		info     *common.BlockchainInfo
//	)
//
//	qscc := system.NewQSCC(client)
//
//	for _, ch := range res.GetChannels() {
//		info, err = qscc.GetChainInfo(ctx, &system.GetChainInfoRequest{ChannelName: ch.ChannelId})
//		if err != nil {
//			return nil, fmt.Errorf("get channel=%s info: %w", ch.GetChannelId(), err)
//		}
//		channels.Channels = append(channels.Channels, &Channel{
//			Name:   ch.GetChannelId(),
//			Height: info.GetHeight(),
//		})
//	}
//
//	return &channels, nil
//}
//
//func (p *PeerService) ListEndorsers(ctx context.Context, _ *empty.Empty) (*Endorsers, error) {
//	var endorsers []string
//
//	client, err := p.Client(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	peersMap := client.PeerPool().GetPeers()
//	if peers, ok := peersMap[client.CurrentIdentity().GetMSPIdentifier()]; ok {
//		for _, mspIDPeer := range peers {
//			endorsers = append(endorsers, mspIDPeer.Uri())
//		}
//	}
//
//	return &Endorsers{Endorsers: endorsers}, nil
//}
//
//func (p *PeerService) CheckChannelConnectivity(ctx context.Context, checkChannel *CheckChannelConnectivityRequest) (*CheckChannelConnectivityResponse, error) {
//	if err := checkChannel.Validate(); err != nil {
//		return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
//
//	p.logger.Debug(`check channel connectivity, request channel data`, zap.String(`channel`, checkChannel.Channel))
//	channelData, err := p.GetChannel(ctx, &GetChannelRequest{
//		ChannelName: checkChannel.Channel,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	var anchorPeers []*channel.Peer
//	for _, org := range channelData.Channel.Organizations {
//		for _, anchorPeer := range org.AnchorPeers {
//			anchorPeers = append(anchorPeers, &channel.Peer{Endpoint: anchorPeer, MspId: org.MspId})
//		}
//	}
//
//	connectivity := &CheckChannelConnectivityResponse{}
//
//	wg := &sync.WaitGroup{}
//	wg.Add(2) // peers and orderer
//
//	ctxConnectivity, cancel := context.WithTimeout(ctx, p.timeouts.CheckConnectivity)
//	defer cancel()
//
//	go func() {
//		p.logger.Debug(`check peers connectivity`, zap.Reflect(`peers`, anchorPeers))
//		connectivity.AnchorPeers = channel.CheckPeersConnectivity(ctxConnectivity, anchorPeers, grpc.WithInsecure()).Peers
//		wg.Done()
//	}()
//
//	go func() {
//		p.logger.Debug(`check orderer connectivity`, zap.Strings(`endpoints`, channelData.Channel.Endpoints))
//		connectivity.Orderer = channel.CheckOrdererConnectivity(ctxConnectivity, channelData.Channel.Endpoints, grpc.WithInsecure())
//		wg.Done()
//	}()
//
//	wg.Wait()
//	return connectivity, nil
//}
//
//func (p *PeerService) GetChannelDiscovery(ctx context.Context, request *CheckChannelConnectivityRequest) (*CheckChannelConnectivityResponse, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (p *PeerService) GetPeerInfo(ctx context.Context, _ *empty.Empty) (*GetPeerInfoRes, error) {
//	channels, err := p.ListChannels(ctx, nil)
//	if err != nil {
//		return nil, err
//	}
//	chaincodes, err := p.GetInstalledChaincodes(ctx, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	fabricVersion, err := p.metricsClient.GetFabricVersion(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	res := &GetPeerInfoRes{
//		MspId:       p.msp.MSPConfig().Name,
//		CertPem:     p.msp.Signer().GetPEM(),
//		CertContent: certToString(p.msp.Signer().GetPEM()),
//		Address:     p.connection.URL,
//		PeerVersion: fabricVersion,
//		Channels:    channels,
//		Chaincodes:  chaincodes,
//	}
//
//	return res, nil
//}
//
//func certToString(certBytes []byte) string {
//	cert, err := x509util.CertificateFromPEM(certBytes)
//	if err != nil {
//		return err.Error()
//	}
//
//	return x509util.CertificateToString(cert)
//}
