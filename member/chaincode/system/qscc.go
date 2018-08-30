package system

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/util"
	qsccPkg "github.com/hyperledger/fabric/core/scc/qscc"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

type qscc struct {
	peer      api.Peer
	identity  msp.SigningIdentity
	processor api.PeerProcessor
}

func (c *qscc) GetChainInfo(ctx context.Context, channelName string) (*common.BlockchainInfo, error) {
	if infoBytes, err := c.endorse(ctx, qsccPkg.GetChainInfo, channelName); err != nil {
		return nil, errors.Wrap(err, `failed to get chainInfo`)
	} else {
		chainInfo := new(common.BlockchainInfo)
		if err = proto.Unmarshal(infoBytes, chainInfo); err != nil {
			return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
		}
		return chainInfo, nil
	}
}

func (c *qscc) GetBlockByNumber(ctx context.Context, channelName string, blockNumber int64) (*common.Block, error) {
	if blockBytes, err := c.endorse(ctx, qsccPkg.GetBlockByNumber, string(blockNumber)); err != nil {
		return nil, errors.Wrap(err, `failed to get block`)
	} else {
		block := new(common.Block)
		if err = proto.Unmarshal(blockBytes, block); err != nil {
			return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
		}
		return block, nil
	}
}

func (c *qscc) GetBlockByHash(ctx context.Context, channelName string, blockHash []byte) (*common.Block, error) {
	if blockBytes, err := c.endorse(ctx, qsccPkg.GetBlockByHash, string(blockHash)); err != nil {
		return nil, errors.Wrap(err, `failed to get block`)
	} else {
		block := new(common.Block)
		if err = proto.Unmarshal(blockBytes, block); err != nil {
			return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
		}
		return block, nil
	}
}

func (c *qscc) GetTransactionByID(ctx context.Context, channelName string, tx api.ChaincodeTx) (*peer.ProcessedTransaction, error) {
	if txBytes, err := c.endorse(ctx, qsccPkg.GetTransactionByID, string(tx)); err != nil {
		return nil, errors.Wrap(err, `failed to get transaction`)
	} else {
		transaction := new(peer.ProcessedTransaction)
		if err = proto.Unmarshal(txBytes, transaction); err != nil {
			return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
		}
		return transaction, nil
	}
}

func (c *qscc) GetBlockByTxID(ctx context.Context, channelName string, tx api.ChaincodeTx) (*common.Block, error) {
	if blockBytes, err := c.endorse(ctx, qsccPkg.GetBlockByTxID, string(tx)); err != nil {
		return nil, errors.Wrap(err, `failed to get block`)
	} else {
		block := new(common.Block)
		if err = proto.Unmarshal(blockBytes, block); err != nil {
			return nil, errors.Wrap(err, `failed to unmarshal protobuf`)
		}
		return block, nil
	}
}
func (c *qscc) endorse(ctx context.Context, fn string, args ...string) ([]byte, error) {
	prop, _, err := c.processor.CreateProposal(&api.DiscoveryChaincode{Name: qsccName, Type: api.CCTypeGoLang}, c.identity, fn, util.ToChaincodeArgs(args...))
	if err != nil {
		return nil, errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.peer.Endorse(ctx, prop)
	if err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
	}
	return resp.Response.Payload, nil
}

func NewQSCC(peer api.Peer, identity msp.SigningIdentity) api.QSCC {
	return &qscc{peer: peer, identity: identity, processor: peerSDK.NewProcessor(``)}
}
