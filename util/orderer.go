package util

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
)

// GetConfigBlockFromOrderer returns config block from orderer by channel name
func GetConfigBlockFromOrderer(ctx context.Context, id msp.SigningIdentity, orderer api.Orderer, channelName string) (*common.Block, error) {
	startPos, endPos := api.SeekNewest()()

	seekEnvelope, err := SeekEnvelope(channelName, startPos, endPos, id)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create seek envelope`)
	}

	lastBlock, err := orderer.Deliver(ctx, seekEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, `failed to fetch last block`)
	}

	blockId, err := protoutil.GetLastConfigIndexFromBlock(lastBlock)
	if err != nil {
		return nil, errors.Wrap(err, `failed to fetch block id with config`)
	}

	startPos, endPos = api.SeekSingle(blockId)()

	seekEnvelope, err = SeekEnvelope(channelName, startPos, endPos, id)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create seek envelope`)
	}

	configBlock, err := orderer.Deliver(ctx, seekEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, `failed to fetch block with config`)
	}

	return configBlock, nil
}
