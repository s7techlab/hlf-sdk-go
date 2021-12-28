package system

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

// These are function names from Invoke first parameter
const (
	JoinChain      string = "JoinChain"
	GetConfigBlock string = "GetConfigBlock"
	GetChannels    string = "GetChannels"
	GetConfigTree  string = `GetConfigTree`
)

type csccV1 struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

func (c *csccV1) JoinChain(ctx context.Context, channelName string, genesisBlock *common.Block) error {
	blockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal block %s", channelName)
	}

	_, err = c.endorse(ctx, JoinChain, string(blockBytes))
	return err
}

func (c *csccV1) GetConfigBlock(ctx context.Context, channelName string) (*common.Block, error) {
	resp, err := c.endorse(ctx, GetConfigBlock, channelName)
	if err != nil {
		return nil, err
	}
	block := new(common.Block)
	if err = proto.Unmarshal(resp, block); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return block, nil
}

func (c *csccV1) GetChannelConfig(ctx context.Context, channelName string) (*common.Config, error) {
	resp, err := c.endorse(ctx, GetConfigTree, channelName)
	if err != nil {
		return nil, err
	}
	configTree := new(peer.ConfigTree)
	if err = proto.Unmarshal(resp, configTree); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return configTree.ChannelConfig, nil
}

func (c *csccV1) GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error) {
	resp, err := c.endorse(ctx, GetChannels)
	if err != nil {
		return nil, err
	}
	channelResp := new(peer.ChannelQueryResponse)
	if err = proto.Unmarshal(resp, channelResp); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return channelResp, nil
}

func (c *csccV1) endorse(ctx context.Context, fn string, args ...string) ([]byte, error) {
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

func NewCSCCV1(peerPool api.PeerPool, identity msp.SigningIdentity) api.CSCC {
	return &csccV1{peerPool: peerPool, identity: identity, processor: peerSDK.NewProcessor(``)}
}
