package transform

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	hlfproto "github.com/s7techlab/hlf-sdk-go/proto"
)

type (
	ActionPayloadTransformer interface {
		Transform(*hlfproto.TransactionAction) error
	}
	ActionPayloadMatch  func(*hlfproto.TransactionAction) bool
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
	if !action.match(txAction) {
		return nil
	}
	for _, mutate := range action.mutators {
		if err := mutate(txAction); err != nil {
			return fmt.Errorf(`kv write mutate: %w`, err)
		}
	}
	return nil
}

func ActionPayloadMatchNil(txAction *hlfproto.TransactionAction) bool {
	return txAction != nil
}

func ActionPayloadMutateProto(target proto.Message) ActionPayloadMutate {
	return func(txAction *hlfproto.TransactionAction) error {
		payloadJSON, err := Proto2JSON(txAction.BytesPayload, target)
		if err != nil {
			return err
		}
		txAction.BytesPayload = payloadJSON
		return nil
	}
}

func ActionPayloadProto(txAction proto.Message) *ActionPayload {
	return NewActionPayload(
		ActionPayloadMatchNil,
		ActionPayloadMutateProto(txAction),
	)
}
