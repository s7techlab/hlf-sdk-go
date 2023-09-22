package transform

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"

	"github.com/s7techlab/hlf-sdk-go/util"
)

type (
	KVWriteTransformer interface {
		Transform(*kvrwset.KVWrite) error
	}
	KVWriteMatch  func(*kvrwset.KVWrite) bool
	KVWriteMutate func(*kvrwset.KVWrite) error

	KVWrite struct {
		match    KVWriteMatch
		mutators []KVWriteMutate
	}
)

func NewKVWrite(match KVWriteMatch, mutators ...KVWriteMutate) *KVWrite {
	return &KVWrite{
		match:    match,
		mutators: mutators,
	}
}

func (kvwrite *KVWrite) Transform(w *kvrwset.KVWrite) error {
	if !kvwrite.match(w) {
		return nil
	}
	for _, mutate := range kvwrite.mutators {
		if err := mutate(w); err != nil {
			return fmt.Errorf(`kv write mutate: %w`, err)
		}
	}
	return nil
}

func KVWriteMatchKeyPrefix(prefixes ...string) KVWriteMatch {
	return func(write *kvrwset.KVWrite) bool {
		keyPrefix, _ := util.SplitCompositeKey(write.Key)
		for _, prefix := range prefixes {
			if keyPrefix == prefix {
				return true
			}
		}

		return false
	}
}

func KVWriteMutateProto(target proto.Message) KVWriteMutate {
	return func(write *kvrwset.KVWrite) error {
		value, err := Proto2JSON(write.Value, target)
		if err != nil {
			return fmt.Errorf(`write mutator key=%s: %w`, write.Key, err)
		}

		write.Value = value
		return nil
	}
}

func KVWriteProtoWithKeyPrefix(prefix string, target proto.Message) *KVWrite {
	return NewKVWrite(
		KVWriteMatchKeyPrefix(prefix),
		KVWriteMutateProto(target),
	)
}

func KVWriteMutatorWithKeyPrefix(prefix string, mutator KVWriteMutate) *KVWrite {
	return NewKVWrite(
		KVWriteMatchKeyPrefix(prefix),
		mutator,
	)
}

func KVWriteKeyObjectTypeReplaceByMap(mapping map[string]string, additionalMutators ...KVWriteMutate) *KVWrite {
	return NewKVWrite(
		KVWriteMatchKeyPrefix(MappingPrefixes(mapping)...),
		append([]KVWriteMutate{KVWriteKeyReplacer(mapping)}, additionalMutators...)...,
	)
}

func KVWriteKeyReplacer(mapping map[string]string) KVWriteMutate {
	return func(write *kvrwset.KVWrite) error {
		prefix, attributes := util.SplitCompositeKey(write.Key)
		if mappedPrefix, ok := mapping[prefix]; ok {
			mappedKey, err := util.CreateCompositeKey(mappedPrefix, attributes)
			if err != nil {
				return fmt.Errorf(`create mapped composite key: %w`, err)
			}
			write.Key = mappedKey
		}
		return nil
	}
}

func KVWriteKeyReplace(mapping map[string]string, additionalMutators ...KVWriteMutate) *KVWrite {
	return NewKVWrite(KVWriteMatchKey(MappingPrefixes(mapping)...), additionalMutators...)
}

func KVWriteMatchKey(contents ...string) KVWriteMatch {
	return func(write *kvrwset.KVWrite) bool {
		for _, content := range contents {
			if strings.Contains(write.Key, content) {
				return true
			}
		}

		return false
	}
}
