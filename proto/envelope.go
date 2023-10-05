package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/util/txflags"
)

func ParseBlockData(blockData [][]byte, txFilter txflags.ValidationFlags) (*BlockData, error) {
	var envelopes []*Envelope
	for i, envelope := range blockData {
		parsedEnvelope, err := ParseEnvelope(envelope, txFilter.Flag(i))
		if err != nil {
			return nil, fmt.Errorf("parse envelope: %w", err)
		}

		envelopes = append(envelopes, parsedEnvelope)
	}

	return &BlockData{
		Envelopes: envelopes,
	}, nil
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

	signatureHeader := &SignatureHeader{}
	sigHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("signature header: %w", err)
	}
	creator, err := protoutil.UnmarshalSerializedIdentity(sigHeader.Creator)
	if err != nil {
		return nil, fmt.Errorf("parse transaction creator: %w", err)
	}
	signatureHeader.Creator = creator
	signatureHeader.Nonce = sigHeader.Nonce

	if channelHeader.TxId == "" {
		protoutil.SetTxID(channelHeader, sigHeader)
	}

	var tx *Transaction
	switch common.HeaderType(channelHeader.Type) {
	case common.HeaderType_ENDORSER_TRANSACTION:
		tx, err = ParseTransaction(payload, common.HeaderType(channelHeader.Type))
		if err != nil {
			return nil, fmt.Errorf(`endorser transaction from envelope: %w`, err)
		}
	case common.HeaderType_CONFIG:
		tx, err = ParseTransaction(payload, common.HeaderType(channelHeader.Type))
		if err != nil {
			return nil, fmt.Errorf(`endorser transaction from envelope: %w`, err)
		}

		ce := &common.ConfigEnvelope{}
		if err = proto.Unmarshal(payload.Data, ce); err != nil {
			return nil, err
		}

		tx.ChannelConfig, err = ParseChannelConfig(*ce.Config)
		if err != nil {
			return nil, fmt.Errorf(`parse channel config: %w`, err)
		}
	}

	return &Envelope{
		Payload: &Payload{
			Header: &Header{
				ChannelHeader:   channelHeader,
				SignatureHeader: signatureHeader,
			},
			Transaction: tx,
		},
		Signature:      envelope.Signature,
		ValidationCode: validationCode,
	}, nil
}
