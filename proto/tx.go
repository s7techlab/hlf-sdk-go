package proto

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

func (x *Transaction) Events() []*peer.ChaincodeEvent {
	var events []*peer.ChaincodeEvent
	for _, a := range x.Actions {
		if a.Payload.Action.ProposalResponsePayload.Extension.Events != nil {
			events = append(events, a.Payload.Action.ProposalResponsePayload.Extension.Events)
		}
	}
	return events
}

func ParseTransaction(payload *common.Payload, transactionType common.HeaderType) (*Transaction, error) {
	var actions []*TransactionAction
	if transactionType == common.HeaderType_ENDORSER_TRANSACTION {
		tx, err := protoutil.UnmarshalTransaction(payload.Data)
		if err != nil {
			return nil, err
		}

		actions, err = ParseTxActions(tx.Actions)
		if err != nil {
			return nil, err
		}
	}

	return &Transaction{
		Actions: actions,
	}, nil
}
