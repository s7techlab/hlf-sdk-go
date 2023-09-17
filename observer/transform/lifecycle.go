package transform

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"

	"github.com/s7techlab/hlf-sdk-go/observer"
)

const (
	LifecycleChaincodeName = "_lifecycle"

	// MetadataPrefix - это префикс ключа стейта, который хранит информацию о ключах
	// в соответствующем неймспейсе. У каждого закомиченного чейнкода в канале есть такой стейт
	MetadataPrefix = "namespaces/metadata"

	// FieldsPrefix - префикс ключа стейтов, которые хранят параметры
	// закомиченного чейнкода в канале
	FieldsPrefix = "namespaces/fields"

	// CollectionField ValidationInfoField EndorsementInfoField SequenceField - суффиксы ключей стейтов,
	// которые хранят параметры закомиченного чейнкода в канале
	CollectionField      = "Collections"
	ValidationInfoField  = "ValidationInfo"
	EndorsementInfoField = "EndorsementInfo"
	SequenceField        = "Sequence"

	Collection      = FieldsPrefix + "/" + CollectionField
	ValidationInfo  = FieldsPrefix + "/" + ValidationInfoField
	EndorsementInfo = FieldsPrefix + "/" + EndorsementInfoField
	Sequence        = FieldsPrefix + "/" + SequenceField

	strByteZero = string(byte(0))
)

func keyReplace(key string) string {
	// lifecycle key is look like 'namespaces/metadata/{chaincode_id}' or 'namespaces/fields/{chaincode_id}/{field}'
	splitKey := strings.Split(key, "/")
	switch splitKey[1] {
	case "metadata":
		// here 3 elements: [namespaces, metadata, {chaincode_id}]
		// make key '{zeroByte}namespaces/metadata/{zeroByte}{chaincode_id}{zeroByte}
		key = fmt.Sprintf("%s%s/%s%s%s%s", strByteZero, splitKey[0], splitKey[1], strByteZero, splitKey[2], strByteZero)

	case "fields":
		// here 4 elements: [namespaces, fields, {chaincode_id}, {field}]
		// make key '{zeroByte}namespaces/fields/{field}{zeroByte}{chaincode_id}{zeroByte}
		key = fmt.Sprintf("%s%s/%s/%s%s%s%s", strByteZero, splitKey[0], splitKey[1], splitKey[3], strByteZero, splitKey[2], strByteZero)
	}

	return key
}

var LifecycleTransformers = []observer.BlockTransformer{
	NewAction(
		TxChaincodeIDMatch(LifecycleChaincodeName),
		WithKVWriteTransformer(
			KVWriteKeyReplace(LifecycleStateKeyStrMapping(), func(write *kvrwset.KVWrite) error {
				write.Key = keyReplace(write.Key)
				return nil
			}),
		),
		WithKVReadTransformer(
			KVReadKeyReplace(LifecycleStateKeyStrMapping(), func(read *kvrwset.KVRead) error {
				read.Key = keyReplace(read.Key)
				return nil
			}),
		),
	),
}

func LifecycleStateKeyStrMapping() map[string]string {
	mapping := make(map[string]string)
	mapping[MetadataPrefix] = strByteZero
	mapping[FieldsPrefix] = strByteZero
	return mapping
}
