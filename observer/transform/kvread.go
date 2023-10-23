package transform

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"

	hlfproto "github.com/s7techlab/hlf-sdk-go/proto"
)

type (
	KVReadTransformer interface {
		Transform(*kvrwset.KVRead) error
	}

	KVReadMatch  func(*kvrwset.KVRead) bool
	KVReadMutate func(*kvrwset.KVRead) error

	KVRead struct {
		match    KVReadMatch
		mutators []KVReadMutate
	}
)

func NewKVRead(match KVReadMatch, mutators ...KVReadMutate) *KVRead {
	return &KVRead{
		match:    match,
		mutators: mutators,
	}
}

func (kvread *KVRead) Transform(w *kvrwset.KVRead) error {
	if !kvread.match(w) {
		return nil
	}
	for _, mutate := range kvread.mutators {
		if err := mutate(w); err != nil {
			return fmt.Errorf(`kv read mutate: %w`, err)
		}
	}
	return nil
}

func KVReadMatchKeyPrefix(prefixes ...string) KVReadMatch {
	return func(read *kvrwset.KVRead) bool {
		keyPrefix, _ := hlfproto.SplitCompositeKey(read.Key)
		for _, prefix := range prefixes {
			if keyPrefix == prefix {
				return true
			}
		}

		return false
	}
}

func KVReadProtoWithKeyPrefix(prefix string, target proto.Message) *KVRead {
	return NewKVRead(
		KVReadMatchKeyPrefix(prefix),
	)
}

func KVReadMutatorWithKeyPrefix(prefix string, mutator KVReadMutate) *KVRead {
	return NewKVRead(
		KVReadMatchKeyPrefix(prefix),
		mutator,
	)
}

func KVReadKeyObjectTypeReplaceByMap(mapping map[string]string, additionalMutators ...KVReadMutate) *KVRead {
	return NewKVRead(
		KVReadMatchKeyPrefix(MappingPrefixes(mapping)...),
		append([]KVReadMutate{KVReadKeyReplacer(mapping)}, additionalMutators...)...,
	)
}

func KVReadKeyReplacer(mapping map[string]string) KVReadMutate {
	return func(read *kvrwset.KVRead) error {
		prefix, attributes := hlfproto.SplitCompositeKey(read.Key)
		if mappedPrefix, ok := mapping[prefix]; ok {
			mappedKey, err := hlfproto.CreateCompositeKey(mappedPrefix, attributes)
			if err != nil {
				return fmt.Errorf(`create mapped composite key: %w`, err)
			}
			read.Key = mappedKey
		}
		return nil
	}
}

func KVReadKeyReplace(mapping map[string]string, additionalMutators ...KVReadMutate) *KVRead {
	return NewKVRead(KVReadMatchKey(MappingPrefixes(mapping)...), additionalMutators...)
}

func KVReadMatchKey(contents ...string) KVReadMatch {
	return func(read *kvrwset.KVRead) bool {
		for _, content := range contents {
			if strings.Contains(read.Key, content) {
				return true
			}
		}

		return false
	}
}
