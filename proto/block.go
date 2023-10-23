package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	bft "github.com/s7techlab/hlf-sdk-go/proto/smartbft"
	bftcommon "github.com/s7techlab/hlf-sdk-go/proto/smartbft/common"
	"github.com/s7techlab/hlf-sdk-go/proto/txflags"
)

type (
	parseBlockOpts struct {
		configBlock *common.Block
	}

	ParseBlockOpt func(*parseBlockOpts)
)

func WithConfigBlock(configBlock *common.Block) ParseBlockOpt {
	return func(opts *parseBlockOpts) {
		opts.configBlock = configBlock
	}
}

func ParseBlock(block *common.Block, opts ...ParseBlockOpt) (*Block, error) {
	var parsingOpts parseBlockOpts
	for _, opt := range opts {
		opt(&parsingOpts)
	}

	var err error
	parsedBlock := &Block{
		Header:   block.Header,
		Data:     &BlockData{},
		Metadata: &BlockMetadata{},
	}

	txFilter := txflags.ValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	if parsedBlock.Data, err = ParseBlockData(block.GetData().GetData(), txFilter); err != nil {
		return nil, fmt.Errorf("parse block data: %w", err)
	}

	// parse Raft orderer identity
	raftOrdererIdentity, err := ParseOrdererIdentity(block)
	if err != nil {
		return nil, fmt.Errorf("parse orderer identity from block: %w", err)
	}

	if raftOrdererIdentity != nil && raftOrdererIdentity.IdBytes != nil {
		parsedBlock.Metadata.OrdererSignatures = append(parsedBlock.Metadata.OrdererSignatures, &OrdererSignature{Identity: raftOrdererIdentity})
	}

	// parse BFT orderer identities, if there is at least one config block was sent
	if parsingOpts.configBlock != nil {
		var bftOrdererIdentities []*OrdererSignature
		bftOrdererIdentities, err = ParseBTFOrderersIdentities(block, parsingOpts.configBlock)
		if err != nil {
			return nil, fmt.Errorf("parse bft orderers identities: %w", err)
		}

		parsedBlock.Metadata.OrdererSignatures = append(parsedBlock.Metadata.OrdererSignatures, bftOrdererIdentities...)
	}

	return parsedBlock, nil
}

func ParseOrdererIdentity(cb *common.Block) (*msp.SerializedIdentity, error) {
	meta, err := protoutil.GetMetadataFromBlock(cb, common.BlockMetadataIndex_SIGNATURES)
	if err != nil {
		return nil, fmt.Errorf("get metadata from block: %w", err)
	}

	// TODO check transaction type? what transactions **definitely** have signatures?
	// some transactions have no signatures
	if len(meta.Signatures) == 0 {
		return nil, nil
	}
	// first signature by orderer
	signatureHeader, err := protoutil.UnmarshalSignatureHeader(meta.Signatures[0].SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling signature header from metadata signature header: %w", err)
	}

	serializedIdentity, err := protoutil.UnmarshalSerializedIdentity(signatureHeader.Creator)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling serialized indentity from signature header: %w", err)
	}

	return serializedIdentity, nil
}

func ParseBTFOrderersIdentities(block *common.Block, configBlock *common.Block) ([]*OrdererSignature, error) {
	bftMeta := &bftcommon.BFTMetadata{}
	if err := proto.Unmarshal(block.Metadata.Metadata[common.BlockMetadataIndex_SIGNATURES], bftMeta); err != nil {
		return nil, fmt.Errorf("unmarshaling bft block metadata from metadata: %w", err)
	}

	lastConfig := common.LastConfig{}
	err := proto.Unmarshal(bftMeta.Value, &lastConfig)
	if err != nil {
		return nil, nil // it shouldn't return error
	}

	configEnvelope, err := createConfigEnvelope(configBlock.Data.Data[0])
	if err != nil {
		return nil, nil // it shouldn't return error
	}

	ct := configEnvelope.Config.ChannelGroup.Groups["Orderer"].Values["ConsensusType"].Value

	consensusType := &orderer.ConsensusType{}
	err = proto.Unmarshal(ct, consensusType)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling consensus type from config envelope concensus type: %w", err)
	}

	configMetadata := &bft.ConfigMetadata{}
	err = proto.Unmarshal(consensusType.Metadata, configMetadata)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling bft config metadata from concensus type metadata: %w", err)
	}

	var ordererSignatures []*OrdererSignature
	for _, consenter := range configMetadata.Consenters {
		var identity msp.SerializedIdentity
		if err = proto.Unmarshal(consenter.Identity, &identity); err != nil {
			return nil, err
		}

		// among all channel orderers, find those that signed this block
		for _, signature := range bftMeta.Signatures {
			if signature.SignerId == consenter.ConsenterId {
				ordererSignatures = append(ordererSignatures, &OrdererSignature{
					Identity:  &identity,
					Signature: signature.Signature,
				})
			}
		}
	}

	return ordererSignatures, nil
}

// createConfigEnvelope creates configuration envelope proto
func createConfigEnvelope(data []byte) (*common.ConfigEnvelope, error) {
	envelope := &common.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling envelope from config block: %w", err)
	}

	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, fmt.Errorf("unmarshaling payload from envelope: %w", err)
	}

	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, fmt.Errorf("unmarshaling channel header from payload: %w", err)
	}

	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG {
		return nil, fmt.Errorf("block must be of type 'CONFIG'")
	}

	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, fmt.Errorf("unmarshaling config envelope from payload: %w", err)
	}

	return configEnvelope, nil
}

func (x *Block) ValidEnvelopes() []*Envelope {
	var envs []*Envelope
	for _, e := range x.Data.Envelopes {
		if e.ValidationCode != peer.TxValidationCode_VALID {
			continue
		}

		envs = append(envs, e)
	}

	return envs
}

func (x *Block) BlockDate() *timestamp.Timestamp {
	var max *timestamp.Timestamp
	for _, envelope := range x.ValidEnvelopes() {
		ts := envelope.GetPayload().GetHeader().GetChannelHeader().GetTimestamp()

		if ts.AsTime().After(max.AsTime()) {
			max = ts
		}
	}
	return max
}
