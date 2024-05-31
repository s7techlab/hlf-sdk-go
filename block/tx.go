package block

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/proto/block"
)

func ParseEndorserTransaction(payload *common.Payload) (*block.Transaction, error) {
	var actions []*block.TransactionAction
	tx, err := protoutil.UnmarshalTransaction(payload.Data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal transaction from payload data: %w", err)
	}

	actions, err = ParseTxActions(tx.Actions)
	if err != nil {
		return nil, fmt.Errorf("parse transaction actions: %w", err)
	}

	return &block.Transaction{
		Actions: actions,
	}, nil
}
