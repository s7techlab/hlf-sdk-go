package proto

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

type (
	TransactionAction struct {
		Event *peer.ChaincodeEvent
	}

	TransactionsActions []*TransactionAction
)

func ParseTxActions(txActions []*peer.TransactionAction) ([]*TransactionAction, error) {
	var parsedTxActions []*TransactionAction

	for _, action := range txActions {
		txAction, err := ParseTxAction(action)
		if err != nil {
			return nil, fmt.Errorf(`tx action: %w`, err)
		}
		parsedTxActions = append(parsedTxActions, txAction)
	}

	return parsedTxActions, nil
}

func ParseTxAction(txAction *peer.TransactionAction) (*TransactionAction, error) {

	ccActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(txAction.Payload)
	if err != nil {
		return nil, fmt.Errorf(`chaincode action payload: %w`, err)
	}

	proposalResponsePayload, err := protoutil.UnmarshalProposalResponsePayload(
		ccActionPayload.Action.ProposalResponsePayload)
	if err != nil {
		return nil, fmt.Errorf(`proposal response payload:  %w`, err)
	}

	chaincodeAction, err := protoutil.UnmarshalChaincodeAction(proposalResponsePayload.Extension)
	if err != nil {
		return nil, fmt.Errorf(`chaincode action from proposal response: %w`, err)
	}

	ccEvent, err := protoutil.UnmarshalChaincodeEvents(chaincodeAction.Events)
	if err != nil {
		return nil, fmt.Errorf(`event from chaincode action: %w`, err)
	}

	parsedTxAction := &TransactionAction{
		Event: ccEvent,
	}

	return parsedTxAction, nil
}
