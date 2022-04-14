package tx

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/proto"
)

type SeekBlock struct {
	Channel string
	Signer  msp.SigningIdentity

	Start *orderer.SeekPosition
	Stop  *orderer.SeekPosition
}

func (sb SeekBlock) CreateEnvelope() (*common.Envelope, error) {
	return NewSeekBlockEnvelope(sb.Channel, sb.Signer, sb.Start, sb.Stop)
}

func NewSeekGenesisEnvelope(channel string, signer msp.SigningIdentity) (*common.Envelope, error) {
	start := proto.NewSeekSpecified(0)
	stop := proto.NewSeekSpecified(0)

	return NewSeekBlockEnvelope(channel, signer, start, stop)
}

func NewSeekBlockEnvelope(channel string, signer msp.SigningIdentity, start, stop *orderer.SeekPosition) (*common.Envelope, error) {

	signerSerialized, err := signer.Serialize()
	if err != nil {
		return nil, fmt.Errorf(`serialize signer: %w`, err)
	}

	txParams, err := GenerateParamsForSerializedIdentity(signerSerialized)
	if err != nil {
		return nil, fmt.Errorf(`tx id: %w`, err)
	}

	seekInfo, err := proto.NewMarshalledSeekInfo(start, stop)
	if err != nil {
		return nil, fmt.Errorf(`seekInfo: %w`, err)
	}

	header, err := proto.NewCommonHeader(common.HeaderType_DELIVER_SEEK_INFO, txParams.ID, txParams.Nonce, txParams.Timestamp, signerSerialized, channel, ``)
	if err != nil {
		return nil, fmt.Errorf(`payload header: %w`, err)
	}

	payload, err := proto.NewMarshalledCommonPayload(header, seekInfo)
	if err != nil {
		return nil, fmt.Errorf(`common payload: %w`, err)
	}

	return proto.NewCommonEnvelope(payload, signer)
}
