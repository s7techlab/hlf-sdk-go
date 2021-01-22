// +build fabric2

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

type cscc struct {
	peerPool  api.PeerPool
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

// These are function names from Invoke first parameter
const (
	JoinChain            string = "JoinChain"
	JoinChainBySnapshot  string = "JoinChainBySnapshot"
	JoinBySnapshotStatus string = "JoinBySnapshotStatus"
	GetConfigBlock       string = "GetConfigBlock"
	GetChannelConfig     string = "GetChannelConfig"
	GetChannels          string = "GetChannels"
)

func (c cscc) JoinChain(ctx context.Context, channelName string, genesisBlock *common.Block) error {
	blockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal block %s", channelName)
	}

	_, err = c.endorse(ctx, JoinChain, string(blockBytes))
	return err
}

func (c cscc) GetConfigBlock(ctx context.Context, channelName string) (*common.Block, error) {
	resp, err := c.endorse(ctx, GetConfigBlock, channelName)
	var block common.Block
	if err = proto.Unmarshal(resp, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &block, nil
}

func (c cscc) GetChannelConfig(ctx context.Context, channelName string) (*common.Config, error) {
	resp, err := c.endorse(ctx, GetChannelConfig, channelName)
	var chConfig common.Config
	if err = proto.Unmarshal(resp, &chConfig); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &chConfig, nil
}

func (c cscc) GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error) {
	resp, err := c.endorse(ctx, GetChannels)
	var peerResp peer.ChannelQueryResponse
	if err = proto.Unmarshal(resp, &peerResp); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &peerResp, nil
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
