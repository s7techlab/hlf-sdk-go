package subs

import (
	"context"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	utilSDK "github.com/s7techlab/hlf-sdk-go/util"
	"google.golang.org/grpc"
)

type eventSubscription struct {
	blockSub  api.BlockSubscription
	ccName    string
	events    chan *peer.ChaincodeEvent
	blockChan chan *common.Block
	errChan   chan error
}

func (es *eventSubscription) Events() (chan *peer.ChaincodeEvent, error) {
	var err error

	if es.blockChan, err = es.blockSub.Blocks(); err != nil {
		return nil, errors.Wrap(err, `failed to initiate block subscription`)
	}

	go es.handleCCSubscription()

	return es.events, nil
}

func (es *eventSubscription) Errors() chan error {
	return es.blockSub.Errors()
}

func (es *eventSubscription) handleCCSubscription() {
	for {
		select {
		case block, ok := <-es.blockChan:
			if !ok {
				return
			}
			txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
			for i, r := range block.Data.Data {
				if txFilter.IsValid(i) {
					if ev, err := utilSDK.GetEventFromEnvelope(r); err != nil {
						es.errChan <- &api.EnvelopeParsingError{Err: err}
					} else {
						if ev != nil {
							es.events <- ev
						}
					}
				} else {
					env, _ := utils.GetEnvelopeFromBlock(r)
					p, _ := utils.GetPayload(env)
					chHeader, _ := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
					es.errChan <- &api.InvalidTxError{TxId: api.ChaincodeTx(chHeader.TxId), Code: txFilter.Flag(i)}
				}
			}
		case err, ok := <-es.blockSub.Errors():
			if !ok {
				return
			}
			es.errChan <- errors.Wrap(err, `block error:`)
		}
	}
}

func (es *eventSubscription) Close() error {
	close(es.errChan)
	return es.blockSub.Close()
}

func NewEventSubscription(ctx context.Context, channelName string, ccName string, identity msp.SigningIdentity, conn *grpc.ClientConn, seekOpt ...api.EventCCSeekOption) api.EventCCSubscription {
	return &eventSubscription{
		ccName:   ccName,
		events:   make(chan *peer.ChaincodeEvent),
		errChan:  make(chan error),
		blockSub: NewBlockSubscription(ctx, channelName, identity, conn, seekOpt...),
	}
}
