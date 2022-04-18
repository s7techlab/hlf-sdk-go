package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/msp"
)

// NewMarshalledCommonPayload returns marshalled payload from headers and data
func NewMarshalledCommonPayload(header *common.Header, data []byte) ([]byte, error) {
	return proto.Marshal(&common.Payload{
		Header: header,
		Data:   data})
}

func NewCommonEnvelope(payload []byte, signer msp.SigningIdentity) (*common.Envelope, error) {
	payloadSignature, err := signer.Sign(payload)
	if err != nil {
		return nil, fmt.Errorf(`sign payloadl: %w`, err)
	}

	return &common.Envelope{
		Payload:   payload,
		Signature: payloadSignature,
	}, nil
}
