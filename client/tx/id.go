package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric/msp"

	"github.com/atomyze-ru/hlf-sdk-go/crypto"
)

type Params struct {
	ID        string
	Nonce     []byte
	Timestamp *timestamp.Timestamp
}

func GenerateParams(creator msp.SigningIdentity) (*Params, error) {
	serialized, err := creator.Serialize()
	if err != nil {
		return nil, fmt.Errorf(`serialize identity: %w`, err)
	}

	return GenerateParamsForSerializedIdentity(serialized)
}

func GenerateParamsForSerializedIdentity(creator []byte) (*Params, error) {
	id, nonce, err := GenerateIDForSerializedIdentity(creator)
	if err != nil {
		return nil, fmt.Errorf(`tx id: %w`, err)
	}

	return &Params{
		ID:        id,
		Nonce:     nonce,
		Timestamp: TimestampNow(),
	}, nil
}

func GenerateID(creator msp.SigningIdentity) (id string, nonce []byte, err error) {
	serialized, err := creator.Serialize()
	if err != nil {
		return ``, nil, fmt.Errorf(`serialize identity: %w`, err)
	}

	return GenerateIDForSerializedIdentity(serialized)
}

func GenerateIDForSerializedIdentity(creator []byte) (id string, nonce []byte, err error) {
	if nonce, err = crypto.RandomBytes(24); err != nil {
		return ``, nil, fmt.Errorf(`get tx nonce: %w`, err)
	}

	f := sha256.New()
	f.Write(append(nonce, creator...))
	return hex.EncodeToString(f.Sum(nil)), nonce, nil
}

func TimestampNow() *timestamp.Timestamp {
	return ptypes.TimestampNow()
}
