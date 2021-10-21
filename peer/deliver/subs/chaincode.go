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
		events:      make(chan *peer.ChaincodeEvent),
	}
}

type EventSubscription struct {
	chaincodeID string
	fromTx      string
	events      chan *peer.ChaincodeEvent

	ErrorCloser
}

//func (e *EventSubscription) Events() <-chan *peer.ChaincodeEvent {
//	return e.events
//}

func (e *EventSubscription) Events() chan *peer.ChaincodeEvent {
	return e.events
}

func (e *EventSubscription) Handler(block *common.Block) bool {
	if block == nil {
		close(e.events)
	} else {
		txFilter := txflags.ValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
		for i, r := range block.GetData().GetData() {
			if txFilter.IsValid(i) {
				ev, err := utilSDK.GetEventFromEnvelope(r)
				if err != nil {
					if utilSDK.IsErrUnsupportedTxType(err) {
						continue
					} else {
						return true
					}
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
				case e.events <- ev:
				case <-e.ErrorCloser.Done():
					return true
				}
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
