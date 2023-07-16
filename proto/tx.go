package proto

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

func ParseTransaction(payload *common.Payload, transactionType common.HeaderType) (*Transaction, error) {
	sigHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("get signature header: %w", err)
	}

	si, err := protoutil.UnmarshalSerializedIdentity(sigHeader.Creator)
	if err != nil {
		return nil, fmt.Errorf("parse transaction creator: %w", err)
	}

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

	parsedTx := &Transaction{
		Actions:         actions,
		CreatorIdentity: si,
	}

	return parsedTx, nil

}

func (x *Transaction) Events() []*peer.ChaincodeEvent {
	var events []*peer.ChaincodeEvent
	for _, a := range x.Actions {
		if a.Event != nil {
			events = append(events, a.Event)
		}
	}
	return events
}
