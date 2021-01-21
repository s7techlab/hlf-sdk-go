package util

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/common/policydsl"
	"github.com/pkg/errors"
)

func GetMSPFromPolicy(policy string) ([]string, error) {
	policyEnvelope, err := policydsl.FromString(policy)
	if err != nil {
		return nil, errors.Wrap(err, `failed to parse policy`)
	}

	mspIds := make([]string, 0)

	for _, id := range policyEnvelope.Identities {
		var mspIdentity msp.SerializedIdentity
		if err = proto.Unmarshal(id.Principal, &mspIdentity); err != nil {
			return nil, errors.Wrap(err, `failed to get MSP identity`)
		} else {
			mspIds = append(mspIds, mspIdentity.Mspid)
		}
	}

	return mspIds, nil
}
