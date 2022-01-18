package subs

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/s7techlab/hlf-sdk-go/v2/util/txflags"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	utilSDK "github.com/s7techlab/hlf-sdk-go/v2/util"
)

func NewEventSubscription(cid string, fromTx api.ChaincodeTx) *EventSubscription {
	return &EventSubscription{
		chaincodeID: cid,
		fromTx:      string(fromTx),
		events: make(chan interface {
			Event() *peer.ChaincodeEvent
			Block() uint64
		}),
	}
}

type ChaincodeEventWithBlock struct {
	event *peer.ChaincodeEvent
	block uint64
}

func (eb *ChaincodeEventWithBlock) Event() *peer.ChaincodeEvent {
	return eb.event
}

func (eb *ChaincodeEventWithBlock) Block() uint64 {
	return eb.block
}

type EventSubscription struct {
	chaincodeID string
	fromTx      string
	events      chan interface {
		Event() *peer.ChaincodeEvent
		Block() uint64
	}

	ErrorCloser
}

func (e *EventSubscription) Events() chan *peer.ChaincodeEvent {
	eventsRaw := make(chan *peer.ChaincodeEvent)
	go func() {
		for {
			event, hasMore := <-e.events
			if !hasMore {
				close(eventsRaw)
				return
			}

			eventsRaw <- event.Event()
		}
	}()
	return eventsRaw
}

func (e *EventSubscription) EventsWithBlock() chan interface {
	Event() *peer.ChaincodeEvent
	Block() uint64
} {
	return e.events
}

func (e *EventSubscription) Handler(block *common.Block) bool {
	if block == nil {
		close(e.events)
		return false
	}

	txFilter := txflags.ValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	for i, r := range block.GetData().GetData() {
		if !txFilter.IsValid(i) {
			continue
		}
		ev, err := utilSDK.GetEventFromEnvelope(r)
		if err != nil {
			if utilSDK.IsErrUnsupportedTxType(err) {
				continue
			}
			return true
		}

		if ev.GetChaincodeId() != e.chaincodeID {
			continue
		}

		if len(e.fromTx) > 0 {
			if ev.TxId != e.fromTx {
				continue
			} else {
				//reset filter and go to next tx from block
				e.fromTx = ``
				continue
			}
		}

		select {
		case e.events <- &ChaincodeEventWithBlock{
			event: ev,
			block: block.Header.Number,
		}:
		case <-e.ErrorCloser.Done():
			return true
		}
	}

	return false
}

func (e *EventSubscription) Serve(base ErrorCloser, readyForHandling ReadyForHandling) *EventSubscription {
	e.ErrorCloser = base
	readyForHandling()
	return e
}
