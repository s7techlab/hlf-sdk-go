package channel

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/proto"
)

func (c *Core) Join(ctx context.Context) error {
	channelGenesis, err := c.getGenesisBlockFromOrderer(ctx)
	if err != nil {
		return errors.Wrap(err, `failed to retrieve genesis block from orderer`)
	}

	// todo: refactor
	peers := c.peerPool.GetMSPPeers(c.mspId)

	if len(peers) == 0 {
		return fmt.Errorf(`no peeers for msp if=%s`, c.mspId)
	}

	cscc := system.NewCSCC(peers[0], proto.FabricVersionIsV2(c.fabricV2))

	_, err = cscc.JoinChain(ctx, &system.JoinChainRequest{
		Channel:      c.chanName,
		GenesisBlock: channelGenesis,
	})

	return err
}

func (c *Core) getGenesisBlockFromOrderer(ctx context.Context) (*common.Block, error) {
	requestBlockEnvelope, err := tx.NewSeekGenesisEnvelope(c.chanName, c.identity)
	if err != nil {
		return nil, fmt.Errorf(`request block envelope: %w`, err)
	}
	return c.orderer.Deliver(ctx, requestBlockEnvelope)
}
