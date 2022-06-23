package subs

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/atomyze-ru/hlf-sdk-go/proto"
)

type ChaincodeEventWithBlock struct {
	event       *peer.ChaincodeEvent
	block       uint64
	txTimestamp *timestamp.Timestamp
}

func (eb *ChaincodeEventWithBlock) Event() *peer.ChaincodeEvent {
	return eb.event
}

func (eb *ChaincodeEventWithBlock) Block() uint64 {
	return eb.block
}

func (eb *ChaincodeEventWithBlock) TxTimestamp() *timestamp.Timestamp {
	return eb.txTimestamp
}

func NewEventSubscription(cid string, fromTxID string) *EventSubscription {
	return &EventSubscription{
		chaincodeID: cid,
		fromTx:      fromTxID,
		events: make(chan interface {
			Event() *peer.ChaincodeEvent
			Block() uint64
			TxTimestamp() *timestamp.Timestamp
		}),
	}
}

type EventSubscription struct {
	chaincodeID string
	fromTx      string
	events      chan interface {
		Event() *peer.ChaincodeEvent
		Block() uint64
		TxTimestamp() *timestamp.Timestamp
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

func (e *EventSubscription) EventsExtended() chan interface {
	Event() *peer.ChaincodeEvent
	Block() uint64
	TxTimestamp() *timestamp.Timestamp
} {
	return e.events
}

func (e *EventSubscription) Handler(block *common.Block) bool {
	if block == nil {
		close(e.events)
		return false
	}

	parsedBlock, err := proto.ParseBlock(block)
	if err != nil {
		return true
	}

	for _, envelope := range parsedBlock.ValidEnvelopes() {

		if envelope.Transaction == nil {
			continue
		}

		for _, ev := range envelope.Transaction.Events() {

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
				event:       ev,
				block:       block.Header.Number,
				txTimestamp: envelope.ChannelHeader.Timestamp,
			}:
			case <-e.ErrorCloser.Done():
				return true
			}
		}
	}

	return false
}

func (e *EventSubscription) Serve(base ErrorCloser, readyForHandling ReadyForHandling) *EventSubscription {
	e.ErrorCloser = base
	readyForHandling()
	return e
}
