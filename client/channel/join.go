package channel

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
)

func (c *Core) Join(ctx context.Context) error {
	channelGenesis, err := c.getGenesisBlockFromOrderer(ctx)
	if err != nil {
		return errors.Wrap(err, `failed to retrieve genesis block from orderer`)
	}

	var cscc api.CSCC

	if c.fabricV2 {
		cscc = system.NewCSCCV2(c.peerPool, c.identity)
	} else {
		cscc = system.NewCSCCV1(c.peerPool, c.identity)
	}

	return cscc.JoinChain(ctx, c.chanName, channelGenesis)
}

func (c *Core) getGenesisBlockFromOrderer(ctx context.Context) (*common.Block, error) {
	requestBlockEnvelope, err := tx.NewSeekGenesisEnvelope(c.chanName, c.identity)
	if err != nil {
		return nil, fmt.Errorf(`request block envelope: %w`, err)
	}
	return c.orderer.Deliver(ctx, requestBlockEnvelope)
}
