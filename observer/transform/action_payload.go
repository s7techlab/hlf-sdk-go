package transform

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type (
	ActionPayloadTransformer interface {
		Transform(*hlfproto.TransactionAction) error
	}
	ActionPayloadMatch  func(string) bool
	ActionPayloadMutate func(*hlfproto.TransactionAction) error

	ActionPayload struct {
		match    ActionPayloadMatch
		mutators []ActionPayloadMutate
	}
)

func NewActionPayload(match ActionPayloadMatch, mutators ...ActionPayloadMutate) *ActionPayload {
	return &ActionPayload{
		match:    match,
		mutators: mutators,
	}
}

func (action *ActionPayload) Transform(txAction *hlfproto.TransactionAction) error {
	args := txAction.ChaincodeSpec().GetInput().GetArgs()
	if len(args) == 0 {
		return nil
	}

	// args[0] save name method
	if action.match(string(args[0])) {
		for _, mutate := range action.mutators {
			if err := mutate(txAction); err != nil {
				return fmt.Errorf(`Action payload mutate: %w`, err)
			}
		}
	}
	return nil
}

func ActionPayloadMatchFunc(str string) ActionPayloadMatch {
	return func(methodName string) bool {
		return methodName == str
	}
}

func ActionPayloadMutateProto(target proto.Message) ActionPayloadMutate {
	return func(txAction *hlfproto.TransactionAction) error {
		if len(txAction.GetBytesPayload()) == 0 {
			return nil
		}

		if string(txAction.GetBytesPayload())[:1] == "{" {
			return nil
		}

		payloadJSON, err := Proto2JSON(txAction.GetBytesPayload(), target)
		if err != nil {
			log.Printf("%s", err)
		}

		txAction.BytesPayload = ReplaceBytesU0000ToNullBytes(payloadJSON)
		return nil
	}
}

func ActionPayloadProto(methodName string, txAction proto.Message) *ActionPayload {
	return NewActionPayload(
		ActionPayloadMatchFunc(methodName),
		ActionPayloadMutateProto(txAction),
	)
}
