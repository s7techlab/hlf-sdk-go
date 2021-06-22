package system

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

type csccV2 struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

// These are function names from Invoke first parameter
const (
	GetChannelConfig     string = "GetChannelConfig"
)

func (c csccV2) JoinChain(ctx context.Context, channelName string, genesisBlock *common.Block) error {
	blockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal block %s", channelName)
	}

	_, err = c.endorse(ctx, JoinChain, string(blockBytes))
	return err
}

func (c csccV2) GetConfigBlock(ctx context.Context, channelName string) (*common.Block, error) {
	resp, err := c.endorse(ctx, GetConfigBlock, channelName)
	if err != nil {
		return nil, err
	}
	var block common.Block
	if err = proto.Unmarshal(resp, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &block, nil
}

func (c csccV2) GetChannelConfig(ctx context.Context, channelName string) (*common.Config, error) {
	resp, err := c.endorse(ctx, GetChannelConfig, channelName)
	if err != nil {
		return nil, err
	}
	var chConfig common.Config
	if err = proto.Unmarshal(resp, &chConfig); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &chConfig, nil
}

func (c csccV2) GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error) {
	resp, err := c.endorse(ctx, GetChannels)
	if err != nil {
		return nil, err
	}
	var peerResp peer.ChannelQueryResponse
	if err = proto.Unmarshal(resp, &peerResp); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &peerResp, nil
}

func (c *csccV2) endorse(ctx context.Context, fn string, args ...string) ([]byte, error) {
	prop, _, err := c.processor.CreateProposal(&api.DiscoveryChaincode{Name: csccName, Type: api.CCTypeGoLang}, c.identity, fn, util.ToChaincodeArgs(args...), nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
	}
	return resp.Response.Payload, nil
}

func NewCSCCV2(peerPool api.PeerPool, identity msp.SigningIdentity) api.CSCC {
	return &csccV2{peerPool: peerPool, identity: identity, processor: peerSDK.NewProcessor(``)}
}
