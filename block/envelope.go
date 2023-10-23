package block

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/block/txflags"
)

func (x *Envelope) ChannelHeader() *common.ChannelHeader {
	return x.GetPayload().GetHeader().GetChannelHeader()
}

func (x *Envelope) SignatureHeader() *SignatureHeader {
	return x.GetPayload().GetHeader().GetSignatureHeader()
}

func (x *Envelope) TxActions() []*TransactionAction {
	return x.GetPayload().GetTransaction().GetActions()
}

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
		return nil, fmt.Errorf("get envelope from block data: %w", err)
	}

	payload, err := protoutil.UnmarshalPayload(envelope.Payload)
	if err != nil {
		return nil, fmt.Errorf("unmarshal payload from envelope: %w", err)
	}

	channelHeader, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("unmarshal channel header from envelope payload: %w", err)
	}

	signatureHeader := &SignatureHeader{}
	sigHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("unmarshal signature header: %w", err)
	}
	creator, err := protoutil.UnmarshalSerializedIdentity(sigHeader.Creator)
	if err != nil {
		return nil, fmt.Errorf("unmarshal envelope creator: %w", err)
	}
	signatureHeader.Creator = creator
	signatureHeader.Nonce = sigHeader.Nonce

	if channelHeader.TxId == "" {
		protoutil.SetTxID(channelHeader, sigHeader)
	}

	tx := &Transaction{}
	var rawUnparsedTransaction []byte
	switch common.HeaderType(channelHeader.Type) {
	case common.HeaderType_CONFIG:
		ce := &common.ConfigEnvelope{}
		if err = proto.Unmarshal(payload.Data, ce); err != nil {
			return nil, fmt.Errorf("unmarshal payload data to config envelope: %w", err)
		}

		tx.ChannelConfig, err = ParseChannelConfig(*ce.Config)
		if err != nil {
			return nil, fmt.Errorf("parse channel config: %w", err)
		}
	case common.HeaderType_ENDORSER_TRANSACTION:
		tx, err = ParseEndorserTransaction(payload)
		if err != nil {
			return nil, fmt.Errorf("parse endorser transaction: %w", err)
		}

	default:
		rawUnparsedTransaction = payload.Data
	}

	return &Envelope{
		Payload: &Payload{
			Header: &Header{
				ChannelHeader:   channelHeader,
				SignatureHeader: signatureHeader,
			},
			Transaction:            tx,
			RawUnparsedTransaction: rawUnparsedTransaction,
		},
		Signature:      envelope.Signature,
		ValidationCode: validationCode,
	}, nil
}
