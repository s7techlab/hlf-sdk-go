package proto

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

type (
	Transaction struct {
		Actions         TransactionsActions    `json:"transaction_actions"`
		CreatorIdentity msp.SerializedIdentity `json:"creator_identity"`
	}

	Transactions []*Transaction
)

func ParseTransaction(payload *common.Payload, transactionType common.HeaderType) (*Transaction, error) {
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
		log.Println("[hlf.ptoto.tx-parser] got error, skipping it. transaction creator is not 'msp.SerializedIdentity': ", string(sigHeader.Creator))
		err = nil
		//return nil, fmt.Errorf("parse transaction creator: %w", err)
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
		CreatorIdentity: *si,
	}

	return parsedTx, nil

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
