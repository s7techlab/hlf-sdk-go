package subs

import (
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"

	utilSDK "github.com/s7techlab/hlf-sdk-go/util"
)

func NewEventSubscription(cid string) *EventSubscription {
	return &EventSubscription{
		chaincodeID: cid,
		events:      make(chan *peer.ChaincodeEvent),
	}
}

type EventSubscription struct {
	chaincodeID string
	events      chan *peer.ChaincodeEvent

	ErrorCloser
}

func (e *EventSubscription) Events() <-chan *peer.ChaincodeEvent {
	return e.events
}

func (e *EventSubscription) Handler(block *common.Block) bool {
	txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	for i, r := range block.GetData().GetData() {
		if txFilter.IsValid(i) {
			ev, err := utilSDK.GetEventFromEnvelope(r)
			if err != nil {
				return true
			}
			if ev.GetChaincodeId() == e.chaincodeID {
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
