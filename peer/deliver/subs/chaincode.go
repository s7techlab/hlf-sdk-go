package subs

import (
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"

	"github.com/s7techlab/hlf-sdk-go/api"
	utilSDK "github.com/s7techlab/hlf-sdk-go/util"
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
		txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
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

				e.events <- ev
			}
		}
	}
	return false
}

func (e *EventSubscription) Serve(base ErrorCloser) *EventSubscription {
	e.ErrorCloser = base
	return e
}
