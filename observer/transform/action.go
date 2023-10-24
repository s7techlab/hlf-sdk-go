package transform

import (
	"fmt"
	"regexp"

	"github.com/mohae/deepcopy"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
	"github.com/s7techlab/hlf-sdk-go/observer"
)

type (
	Action struct {
		match                     TxActionMatch
		inputArgsTransformers     []InputArgsTransformer
		kvWriteTransformers       []KVWriteTransformer
		kvReadTransformers        []KVReadTransformer
		eventTransformers         []EventTransformer
		actionPayloadTransformers []ActionPayloadTransformer
	}

	ActionOpt func(*Action)

	TxActionMatch  func(*hlfproto.TransactionAction) bool
	TxActionMutate func(*hlfproto.TransactionAction)
)

func WithInputArgsTransformer(inputArgsTransformers ...InputArgsTransformer) ActionOpt {
	return func(a *Action) {
		a.inputArgsTransformers = inputArgsTransformers
	}
}

func WithKVWriteTransformer(kvWriteTransformers ...KVWriteTransformer) ActionOpt {
	return func(a *Action) {
		a.kvWriteTransformers = kvWriteTransformers
	}
}

func WithKVReadTransformer(kvReadTransformers ...KVReadTransformer) ActionOpt {
	return func(a *Action) {
		a.kvReadTransformers = kvReadTransformers
	}
}

func WithEventTransformer(eventTransformers ...EventTransformer) ActionOpt {
	return func(a *Action) {
		a.eventTransformers = eventTransformers
	}
}

func WithActionPayloadTransformer(actionTransformers ...ActionPayloadTransformer) ActionOpt {
	return func(a *Action) {
		a.actionPayloadTransformers = actionTransformers
	}
}

func NewAction(actionMach TxActionMatch, opts ...ActionOpt) *Action {
	a := &Action{
		match: actionMach,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

func (s *Action) Transform(block *observer.ParsedBlock) error {
	if block.Block == nil {
		return nil
	}

	// if block is transformed, copy of block will be saved to block.BlockOriginal
	blockCopy := deepcopy.Copy(block.Block).(*hlfproto.Block)
	blockIsTransformed := false

	for _, envelope := range block.Block.Data.Envelopes {
		if envelope.Payload.Transaction == nil {
			continue
		}

		for _, txAction := range envelope.Payload.Transaction.Actions {
			if !s.match(txAction) {
				continue
			}

			for _, argsTransformer := range s.inputArgsTransformers {
				if err := argsTransformer.Transform(txAction.ChaincodeSpec().Input.Args); err != nil {
					return fmt.Errorf(`transform input args: %w`, err)
				}
			}

			for _, eventTransformer := range s.eventTransformers {
				if err := eventTransformer.Transform(txAction.Event()); err != nil {
					return fmt.Errorf(`transform event: %w`, err)
				}
			}

			for _, rwSet := range txAction.NsReadWriteSet() {
				for _, write := range rwSet.Rwset.Writes {
					for _, kvWriteTransformer := range s.kvWriteTransformers {
						origKey := write.Key
						if err := kvWriteTransformer.Transform(write); err != nil {
							return fmt.Errorf(`transform KV write with key: %s: %w`, write.Key, err)
						}

						if origKey != write.Key {
							blockIsTransformed = true
						}
					}
				}

				for _, read := range rwSet.Rwset.Reads {
					for _, kvReadTransform := range s.kvReadTransformers {
						origKey := read.Key
						if err := kvReadTransform.Transform(read); err != nil {
							return fmt.Errorf(`transform KV read with key: %s: %w`, read.Key, err)
						}
						if origKey != read.Key {
							blockIsTransformed = true
						}
					}
				}

				for _, actionPayloadTransform := range s.actionPayloadTransformers {
					if err := actionPayloadTransform.Transform(txAction); err != nil {
						return fmt.Errorf(`transform action payload: %w`, err)
					}
				}
			}
		}
	}

	if blockIsTransformed {
		block.BlockOriginal = blockCopy
	}

	return nil
}

func TxChaincodeIDMatch(chaincode string) TxActionMatch {
	return func(action *hlfproto.TransactionAction) bool {
		return action.ChaincodeSpec().ChaincodeId.Name == chaincode
	}
}
func TxChaincodesIDMatch(chaincodes ...string) TxActionMatch {
	return func(action *hlfproto.TransactionAction) bool {
		for k := range chaincodes {
			if action.ChaincodeSpec().ChaincodeId.Name == chaincodes[k] {
				return true
			}
		}
		return false
	}
}

func TxChaincodesIDRegexp(chaincodePattern string) TxActionMatch {
	return func(action *hlfproto.TransactionAction) bool {
		matched, _ := regexp.MatchString(chaincodePattern, action.ChaincodeSpec().ChaincodeId.Name)

		return matched
	}
}
func TxChaincodesIDRegexpExclude(chaincodePattern string) TxActionMatch {
	return func(action *hlfproto.TransactionAction) bool {
		matched, _ := regexp.MatchString(chaincodePattern, action.ChaincodeSpec().ChaincodeId.Name)

		return !matched
	}
}

func TxChaincodePatternsIDRegexpExclude(chaincodePatterns ...string) TxActionMatch {
	return func(action *hlfproto.TransactionAction) bool {
		isInSlice := false
		for key := range chaincodePatterns {
			matched, _ := regexp.MatchString(chaincodePatterns[key], action.ChaincodeSpec().ChaincodeId.Name)
			if matched {
				isInSlice = true
			}
		}
		return !isInSlice
	}
}

func TxChaincodeAnyMatch() TxActionMatch {
	return func(action *hlfproto.TransactionAction) bool {
		return true
	}
}
