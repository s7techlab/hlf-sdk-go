package util

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

type ErrUnsupportedTxType struct {
	Type string
}

func (e *ErrUnsupportedTxType) Error() string {
	return fmt.Sprintf("err unknown tx type: %s", e.Type)
}

func IsErrUnsupportedTxType(err error) bool {
	switch err.(type) {
	case *ErrUnsupportedTxType:
		return true
	default:
		return false
	}
}

// GetEventFromEnvelope
// Deprecated: use proto.ParseBlock
func GetEventFromEnvelope(envelopeData []byte) (*peer.ChaincodeEvent, error) {
	if envelopeData == nil {
		return nil, errors.New(`no envelope data`)
	}
	if envelope, err := protoutil.GetEnvelopeFromBlock(envelopeData); err != nil {
		return nil, errors.Wrap(err, `failed to get envelope`)
	} else {
		if payload, err := protoutil.UnmarshalPayload(envelope.Payload); err != nil {
			return nil, errors.Wrap(err, `failed to get payload from envelope`)
		} else {
			if channelHeader, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader); err != nil {
				return nil, errors.Wrap(err, `failed to unmarshal channel header`)
			} else {
				switch common.HeaderType(channelHeader.Type) {
				case common.HeaderType_ENDORSER_TRANSACTION:
					if tx, err := protoutil.UnmarshalTransaction(payload.Data); err != nil {
						return nil, errors.Wrap(err, `failed to get transaction`)
					} else {
						if ccActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload); err != nil {
							return nil, errors.Wrap(err, `failed to get chaincode action payload`)
						} else {
							if propRespPayload, err := protoutil.UnmarshalProposalResponsePayload(ccActionPayload.Action.ProposalResponsePayload); err != nil {
								return nil, errors.Wrap(err, `failed to get proposal response payload`)
							} else {
								if caPayload, err := protoutil.UnmarshalChaincodeAction(propRespPayload.Extension); err != nil {
									return nil, errors.Wrap(err, `failed to get chaincode action`)
								} else {
									if ccEvent, err := protoutil.UnmarshalChaincodeEvents(caPayload.Events); err != nil {
										return nil, errors.Wrap(err, `failed to get events`)
									} else {
										return ccEvent, nil
									}
								}
							}
						}
					}
				default:
					return nil, &ErrUnsupportedTxType{
						Type: common.HeaderType_name[channelHeader.Type],
					}
				}
			}
		}
	}
}
