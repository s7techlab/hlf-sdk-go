package subs

import (
	"context"

	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/s7techlab/hlf-sdk-go/api"
	utilSDK "github.com/s7techlab/hlf-sdk-go/util"
	"go.uber.org/zap"
)

type eventSubscription struct {
	log       *zap.Logger
	blockChan chan *common.Block
	eventChan chan *peer.ChaincodeEvent
	errChan   chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

func (es *eventSubscription) Events() chan *peer.ChaincodeEvent {
	return es.eventChan
}

func (es *eventSubscription) Errors() chan error {
	return es.errChan
}

func (es *eventSubscription) handleCCSubscription() {
	defer func() {
		close(es.eventChan)
		close(es.errChan)
	}()
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
							es.eventChan <- ev
						}
					}
				} else {
					env, _ := utils.GetEnvelopeFromBlock(r)
					p, _ := utils.GetPayload(env)
					chHeader, _ := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
					es.errChan <- &api.InvalidTxError{TxId: api.ChaincodeTx(chHeader.TxId), Code: txFilter.Flag(i)}
				}
			}
		case <-es.ctx.Done():
			return
		}
	}
}

func (es *eventSubscription) Close() error {
	es.cancel()
	return nil
}

func NewEventSubscription(ctx context.Context, blockChan chan *common.Block, log *zap.Logger) api.EventCCSubscription {
	l := log.Named(`EventSubscription`)
	newCtx, cancel := context.WithCancel(ctx)
	return &eventSubscription{
		log:       l,
		eventChan: make(chan *peer.ChaincodeEvent),
		errChan:   make(chan error),
		blockChan: blockChan,
		ctx:       newCtx,
		cancel:    cancel,
	}
}
