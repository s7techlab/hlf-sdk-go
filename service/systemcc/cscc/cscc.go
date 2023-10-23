package cscc

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/s7techlab/hlf-sdk-go/api"
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/channel"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/service"
)

//go:embed cscc.swagger.json
var Swagger []byte

type (
	Service struct {
		UnimplementedCSCCServiceServer

		Querier           *tx.ProtoQuerier
		ChannelListGetter api.ChannelListGetter
		FabricVersion     hlfproto.FabricVersion
	}
)

func FromClient(client api.Client) *Service {
	return NewCSCC(client, hlfproto.FabricVersionIsV2(client.FabricV2()))
}

func NewCSCC(querier api.Querier, version hlfproto.FabricVersion) *Service {
	return &Service{
		// Channel and chaincode are fixed in queries to CSCC
		Querier:           tx.NewProtoQuerier(querier, ``, chaincode.CSCC),
		ChannelListGetter: channel.NewCSCCListGetter(querier),
		FabricVersion:     version,
	}
}

func (c *Service) ServiceDef() *service.Def {
	return service.NewDef(
		_CSCCService_serviceDesc.ServiceName,
		Swagger,
		&_CSCCService_serviceDesc,
		c,
		RegisterCSCCServiceHandlerFromEndpoint,
	)
}

func (c *Service) GetChannels(ctx context.Context, _ *empty.Empty) (*peer.ChannelQueryResponse, error) {
	return c.ChannelListGetter.GetChannels(ctx)
}

func (c *Service) JoinChain(ctx context.Context, request *JoinChainRequest) (*empty.Empty, error) {
	if _, err := c.Querier.Query(ctx, chaincode.CSCCJoinChain, request.GenesisBlock); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (c *Service) GetConfigBlock(ctx context.Context, request *GetConfigBlockRequest) (*common.Block, error) {
	res, err := c.Querier.QueryProto(ctx, []interface{}{chaincode.CSCCGetConfigBlock, request.Channel}, &common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (c *Service) GetChannelConfig(ctx context.Context, request *GetChannelConfigRequest) (*common.Config, error) {
	switch c.FabricVersion {

	case hlfproto.FabricV1:
		res, err := c.Querier.QueryStringsProto(ctx, []string{chaincode.CSCCGetConfigTree, request.Channel}, &peer.ConfigTree{})
		if err != nil {
			return nil, err
		}
		return res.(*peer.ConfigTree).ChannelConfig, nil

	case hlfproto.FabricV2:

		res, err := c.Querier.QueryStringsProto(ctx, []string{chaincode.CSCCGetChannelConfig, request.Channel}, &common.Config{})
		if err != nil {
			return nil, err
		}
		return res.(*common.Config), nil

	default:
		return nil, fmt.Errorf(`fabric version=%s: %w`, c.FabricVersion, hlfproto.ErrUnknownFabricVersion)
	}
}
