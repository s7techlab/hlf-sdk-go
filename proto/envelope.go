package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/v2/util/txflags"
)

type (
	Envelope struct {
		Signature      []byte                `json:"signature"`
		ChannelHeader  *common.ChannelHeader `json:"channel_header"`
		ValidationCode peer.TxValidationCode `json:"validation_code"`
		Transaction    *Transaction          `json:"transaction,omitempty"`
		ChannelConfig  *ChannelConfig        `json:"channel_config,omitempty"`
	}

	Envelopes []*Envelope
)

func ParseEnvelopes(blockData [][]byte, txFilter txflags.ValidationFlags) ([]*Envelope, error) {
	var envelopes []*Envelope
	for i, envelope := range blockData {
		parsedEnvelope, err := ParseEnvelope(envelope, txFilter.Flag(i))
		if err != nil {
			return nil, fmt.Errorf("parse envelope: %w", err)
		}

		envelopes = append(envelopes, parsedEnvelope)
	}

	return envelopes, nil
}

func ParseEnvelope(envelopeData []byte, validationCode peer.TxValidationCode) (*Envelope, error) {
	envelope, err := protoutil.GetEnvelopeFromBlock(envelopeData)
	if err != nil {
		return nil, fmt.Errorf(`envelope from block data: %w`, err)
	}

	payload, err := protoutil.UnmarshalPayload(envelope.Payload)
	if err != nil {
		return nil, fmt.Errorf(`payload from envelope: %w`, err)
	}

	channelHeader, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf(`channel header from envelope payload: %w`, err)
	}

	parsedEnvelope := &Envelope{
		Signature:      envelope.Signature,
		ChannelHeader:  channelHeader,
		ValidationCode: validationCode,
	}

	switch common.HeaderType(channelHeader.Type) {
	case common.HeaderType_ENDORSER_TRANSACTION:
		parsedEnvelope.Transaction, err = ParseEndorserTx(payload.Data)
		if err != nil {
			return nil, fmt.Errorf(`endorser transaction from envelope: %w`, err)
		}
	case common.HeaderType_CONFIG:
		ce := &common.ConfigEnvelope{}
		if err := proto.Unmarshal(payload.Data, ce); err != nil {
			return nil, err
		}

		parsedEnvelope.ChannelConfig, err = ParseChannelConfig(*ce.Config)
		if err != nil {
			return nil, fmt.Errorf(`parse channel config: %w`, err)
		}
	}

	return parsedEnvelope, nil
}
