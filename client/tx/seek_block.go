package tx

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/block"
)

type SeekBlock struct {
	Channel string
	Signer  msp.SigningIdentity

	Start *orderer.SeekPosition
	Stop  *orderer.SeekPosition
}

func (sb SeekBlock) CreateEnvelope() (*common.Envelope, error) {
	return NewSeekBlockEnvelope(sb.Channel, sb.Signer, sb.Start, sb.Stop, nil)
}

func NewSeekGenesisEnvelope(channel string, signer msp.SigningIdentity, tlsCertHash []byte) (*common.Envelope, error) {
	start := block.NewSeekSpecified(0)
	stop := block.NewSeekSpecified(0)

	return NewSeekBlockEnvelope(channel, signer, start, stop, tlsCertHash)
}

func NewSeekBlockEnvelope(channel string, signer msp.SigningIdentity, start, stop *orderer.SeekPosition, tlsCertHash []byte) (
	*common.Envelope, error) {
	if signer == nil {
		return nil, errors.New(`signer should be defined`)
	}
	signerSerialized, err := signer.Serialize()
	if err != nil {
		return nil, fmt.Errorf(`serialize signer: %w`, err)
	}

	txParams, err := GenerateParamsForSerializedIdentity(signerSerialized)
	if err != nil {
		return nil, fmt.Errorf(`tx id: %w`, err)
	}

	seekInfo, err := block.NewMarshalledSeekInfo(start, stop)
	if err != nil {
		return nil, fmt.Errorf(`seekInfo: %w`, err)
	}

	header, err := block.NewCommonHeader(
		common.HeaderType_DELIVER_SEEK_INFO,
		txParams.ID,
		txParams.Nonce,
		txParams.Timestamp,
		signerSerialized,
		channel, ``,
		tlsCertHash)
	if err != nil {
		return nil, fmt.Errorf(`payload header: %w`, err)
	}

	payload, err := block.NewMarshalledCommonPayload(header, seekInfo)
	if err != nil {
		return nil, fmt.Errorf(`common payload: %w`, err)
	}

	return block.NewCommonEnvelope(payload, signer)
}
