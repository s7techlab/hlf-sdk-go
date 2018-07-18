package system

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/util"
	csccPkg "github.com/hyperledger/fabric/core/scc/cscc"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)


type cscc struct {
	peer      api.Peer
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

func (c *cscc) JoinChain(channelName string, genesisBlock *common.Block) error {
	blockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return errors.Wrap(err, `failed to marshal block`)
	}

	_, err = c.endorse(csccPkg.JoinChain, channelName, string(blockBytes))
	return err
}

func (c *cscc) GetConfigBlock(channelName string) (*common.Block, error) {
	resp, err := c.endorse(csccPkg.GetConfigBlock, channelName)
	if err != nil {
		return nil, err
	}
	block := new(common.Block)
	if err = proto.Unmarshal(resp, block); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return block, nil
}

func (c *cscc) GetConfigTree(channelName string) (*peer.ConfigTree, error) {
	resp, err := c.endorse(csccPkg.GetConfigTree, channelName)
	if err != nil {
		return nil, err
	}
	configTree := new(peer.ConfigTree)
	if err = proto.Unmarshal(resp, configTree); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return configTree, nil
}

func (c *cscc) Channels() (*peer.ChannelQueryResponse, error) {
	resp, err := c.endorse(csccPkg.GetChannels)
	if err != nil {
		return nil, err
	}
	channelResp := new(peer.ChannelQueryResponse)
	if err = proto.Unmarshal(resp, channelResp); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
	}
	return channelResp, nil
}

func (c *cscc) endorse(fn string, args ...string) ([]byte, error) {
	prop, _, err := c.processor.CreateProposal(&api.DiscoveryChaincode{Name: csccName, Type: api.CCTypeGoLang}, c.identity, fn, util.ToChaincodeArgs(args...))
	if err != nil {
		return nil, errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.peer.Endorse(prop)
	if err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
	}
	return resp.Response.Payload, nil
}

func NewCSCC(peer api.Peer, identity msp.SigningIdentity) api.CSCC {
	return &cscc{peer: peer, identity: identity, processor: peerSDK.NewProcessor(``)}
}
