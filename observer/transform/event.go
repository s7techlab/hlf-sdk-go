package transform

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type (
	EventTransformer interface {
		Transform(*peer.ChaincodeEvent) error
	}
	EventMatch  func(string) bool
	EventMutate func(*peer.ChaincodeEvent) error

	Event struct {
		match  EventMatch
		mutate EventMutate
	}
)

func NewEvent(match EventMatch, mutate EventMutate) *Event {
	return &Event{
		match:  match,
		mutate: mutate,
	}
}

func (e *Event) Transform(event *peer.ChaincodeEvent) error {
	if e.match(event.EventName) {
		return e.mutate(event)
	}
	return nil
}

func EventProto(eventName string, target proto.Message) *Event {
	return NewEvent(
		EventMatchFunc(eventName),
		EventMutateProto(target),
	)
}

func EventMatchFunc(str string) EventMatch {
	return func(eventName string) bool {
		return eventName == str
	}
}

func EventMutateProto(target proto.Message) EventMutate {
	return func(event *peer.ChaincodeEvent) error {
		payloadJSON, err := Proto2JSON(event.Payload, target)
		if err != nil {
			return fmt.Errorf(`event payload mutator: %w`, err)
		}

		event.Payload = payloadJSON
		return nil
	}
}
