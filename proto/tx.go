package proto

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

type (
	Transaction struct {
		Actions         TransactionsActions    `json:"transaction_actions"`
		CreatorIdentity msp.SerializedIdentity `json:"creator_identity"`
		Signature       []byte                 `json:"signature"`
	}

	Transactions []*Transaction
)

func ParseEndorserTx(envelopePayloadBytes []byte) (*Transaction, error) {
	tx, err := protoutil.UnmarshalTransaction(envelopePayloadBytes)
	if err != nil {
		return nil, err
	}

	parsedTx := &Transaction{}
	parsedTx.Actions, err = ParseTxActions(tx.Actions)
	if err != nil {
		return nil, err
	}

	payload, err := protoutil.UnmarshalPayload(envelopePayloadBytes)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get payload from envelope`)
	}

	sigHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature header: %w", err)
	}

	si, err := protoutil.UnmarshalSerializedIdentity(sigHeader.Creator)
	if err != nil {
		// in some transactions we get some unknown proto message with chaincodes(!?), dont know how to deal with it now
		// ---- example
		// �
		// �
		// _lifecycle�
		// "ApproveChaincodeDefinitionForMyOrg
		// ^basic1.0JNL
		// Jbasic_1.0:770d76a4369f9121d4945c6782dd42c4db7c130c3c7b77b73d9894414b5a3da9
		// log.Println("[hlf.ptoto.tx-parser] got error, skipping it. transaction creator is not 'msp.SerializedIdentity': ", string(sigHeader.Creator))
		err = nil
		//return nil, fmt.Errorf("parse transaction creator: %w", err)
	}
	parsedTx.CreatorIdentity = *si

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
