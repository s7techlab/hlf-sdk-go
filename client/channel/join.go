package channel

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/util"
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

	return cscc.JoinChain(ctx, c.name, channelGenesis)
}

func (c *Core) getGenesisBlockFromOrderer(ctx context.Context) (*common.Block, error) {
	ordererSeekInfo := &orderer.SeekInfo{
		Start:    &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: 0}}},
		Stop:     &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: 0}}},
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}

	seekBytes, err := proto.Marshal(ordererSeekInfo)
	if err != nil {
		return nil, errors.Wrap(err, `failed to marshal seekInfo bytes`)
	}

	txId, nonce, err := util.NewTxWithNonce(c.identity)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get new txId`)
	}

	chHeader, err := util.NewChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, txId, c.name, 0, nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get channel header`)
	}

	sigHeader, err := util.NewSignatureHeader(c.identity, nonce)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get signature header`)
	}

	payload, err := util.NewPayloadFromHeader(chHeader, sigHeader, seekBytes)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get payload`)
	}

	payloadSignature, err := c.identity.Sign(payload)
	if err != nil {
		return nil, errors.Wrap(err, `failed to sign payload`)
	}
	return c.orderer.Deliver(ctx, &common.Envelope{Payload: payload, Signature: payloadSignature})
}
