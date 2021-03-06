// +build !fabric2

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

type cscc struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

func (c *cscc) JoinChain(ctx context.Context, channelName string, genesisBlock *common.Block) error {
	blockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal block %s", channelName)
	}

	_, err = c.endorse(ctx, JoinChain, string(blockBytes))
	return err
}

func (c *cscc) GetConfigBlock(ctx context.Context, channelName string) (*common.Block, error) {
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

func (c *cscc) GetConfigTree(ctx context.Context, channelName string) (*peer.ConfigTree, error) {
	resp, err := c.endorse(ctx, GetConfigTree, channelName)
	if err != nil {
		return nil, err
	}
	configTree := new(peer.ConfigTree)
	if err = proto.Unmarshal(resp, configTree); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return configTree, nil
}

func (c *cscc) Channels(ctx context.Context) (*peer.ChannelQueryResponse, error) {
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

func (c *cscc) endorse(ctx context.Context, fn string, args ...string) ([]byte, error) {
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

func NewCSCC(peerPool api.PeerPool, identity msp.SigningIdentity) api.CSCC {
	return &cscc{peerPool: peerPool, identity: identity, processor: peerSDK.NewProcessor(``)}
}
