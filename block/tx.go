package block

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

func (x *Transaction) Events() []*peer.ChaincodeEvent {
	var events []*peer.ChaincodeEvent
	for _, a := range x.Actions {
		event := a.GetPayload().GetAction().GetProposalResponsePayload().GetExtension().GetEvents()
		if event != nil {
			events = append(events, event)
		}
	}
	return events
}

func ParseEndorserTransaction(payload *common.Payload) (*Transaction, error) {
	var actions []*TransactionAction
	tx, err := protoutil.UnmarshalTransaction(payload.Data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal transaction from payload data: %w", err)
	}

	actions, err = ParseTxActions(tx.Actions)
	if err != nil {
		return nil, fmt.Errorf("parse transaction actions: %w", err)
	}

	return &Transaction{
		Actions: actions,
	}, nil
}
