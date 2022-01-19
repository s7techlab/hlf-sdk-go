package proto

import (
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

type (
	Transaction struct {
		Actions TransactionsActions
	}

	Transactions []*Transaction
)

func ParseEndorserTx(envelopePayloadData []byte) (*Transaction, error) {
	tx, err := protoutil.UnmarshalTransaction(envelopePayloadData)
	if err != nil {
		return nil, err
	}

	parsedTx := &Transaction{}
	parsedTx.Actions, err = ParseTxActions(tx.Actions)
	if err != nil {
		return nil, err
	}

	return parsedTx, err
}

func (t *Transaction) Events() []*peer.ChaincodeEvent {
	var events []*peer.ChaincodeEvent
	for _, a := range t.Actions {
		if a.Event != nil {
			events = append(events, a.Event)
		}
	}
	return events
}
