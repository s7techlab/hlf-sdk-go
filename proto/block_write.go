package proto

import (
	"errors"
	"regexp"
	"strings"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
)

var (
	ErrEmptyKeyAttrs = errors.New(`empty key attrs`)
)

type (
	Write struct {
		KWWrite *kvrwset.KVWrite

		KeyObjectType string
		KeyAttrs      []string
		Key           string

		Block            uint64
		Chaincode        string
		ChaincodeVersion string
		Tx               string
		Timestamp        *timestamp.Timestamp
	}
)

func (bw *Write) HasKeyObjectType(objectTypes ...string) bool {
	for _, t := range objectTypes {
		if bw.KeyObjectType == t {
			return true
		}
	}
	return false
}

func (bw *Write) HasKeyPartObjectType(objectTypes ...string) bool {
	for _, t := range objectTypes {
		if strings.Contains(bw.KeyObjectType, t) {
			return true
		}
	}
	return false
}

func (bw *Write) HasKeyObjectTypeRegexp(pattern string) bool {
	matched, _ := regexp.MatchString(pattern, bw.KeyObjectType)

	return matched
}
