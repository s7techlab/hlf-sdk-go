package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
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

	parsedTx.CreatorIdentity, err = ParseTransactionCreator(sigHeader)
	if err != nil {
		return nil, fmt.Errorf("parse transaction creator: %w", err)
	}

	return parsedTx, err
}

func ParseTransactionCreator(sigHeader *common.SignatureHeader) (msp.SerializedIdentity, error) {
	creatorIdentity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(sigHeader.Creator, creatorIdentity); err != nil {
		return msp.SerializedIdentity{}, fmt.Errorf("failed to get creator identity: %w", err)
	}
	return *creatorIdentity, nil
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
