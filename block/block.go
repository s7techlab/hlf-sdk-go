package block

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/block/txflags"
	"github.com/s7techlab/hlf-sdk-go/proto/block"
	bft "github.com/s7techlab/hlf-sdk-go/proto/block/smartbft"
	bftcommon "github.com/s7techlab/hlf-sdk-go/proto/block/smartbft/common"
)

var (
	ErrNilBlock       = errors.New("nil block")
	ErrNilConfigBlock = errors.New("nil config block")
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

func ParseBlock(commonBlock *common.Block, opts ...ParseBlockOpt) (*block.Block, error) {
	var parsingOpts parseBlockOpts
	for _, opt := range opts {
		opt(&parsingOpts)
	}

	var err error
	parsedBlock := &block.Block{
		Header:   commonBlock.Header,
		Data:     &block.BlockData{},
		Metadata: &block.BlockMetadata{},
	}

	txFilter := txflags.ValidationFlags(commonBlock.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	if parsedBlock.Data, err = ParseBlockData(commonBlock.GetData().GetData(), txFilter); err != nil {
		return nil, fmt.Errorf("parse block data: %w", err)
	}

	// parse Raft orderer identity
	raftOrdererIdentity, err := ParseOrdererIdentity(commonBlock)
	if err != nil {
		return nil, fmt.Errorf("parse orderer identity from block: %w", err)
	}

	if raftOrdererIdentity != nil && raftOrdererIdentity.IdBytes != nil {
		parsedBlock.Metadata.OrdererSignatures = append(parsedBlock.Metadata.OrdererSignatures,
			&block.OrdererSignature{Identity: raftOrdererIdentity})
	}

	// parse BFT orderer identities, if there is at least one config block was sent
	if parsingOpts.configBlock != nil {
		var bftOrdererIdentities []*block.OrdererSignature
		bftOrdererIdentities, err = ParseBTFOrderersIdentities(commonBlock, parsingOpts.configBlock)
		if err != nil {
			return nil, fmt.Errorf("parse bft orderers identities: %w", err)
		}

		parsedBlock.Metadata.OrdererSignatures = append(parsedBlock.Metadata.OrdererSignatures, bftOrdererIdentities...)
	}

	return parsedBlock, nil
}

func ParseOrdererIdentity(cb *common.Block) (*msp.SerializedIdentity, error) {
	if cb == nil {
		return nil, ErrNilBlock
	}

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

func ParseBTFOrderersIdentities(commonBlock *common.Block, configBlock *common.Block) ([]*block.OrdererSignature, error) {
	if commonBlock == nil {
		return nil, ErrNilBlock
	}

	if configBlock == nil {
		return nil, ErrNilConfigBlock
	}

	bftMeta := &bftcommon.BFTMetadata{}
	if err := proto.Unmarshal(commonBlock.Metadata.Metadata[common.BlockMetadataIndex_SIGNATURES], bftMeta); err != nil {
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

	var ordererSignatures []*block.OrdererSignature
	for _, consenter := range configMetadata.Consenters {
		var identity msp.SerializedIdentity
		if err = proto.Unmarshal(consenter.Identity, &identity); err != nil {
			return nil, err
		}

		// among all channel orderers, find those that signed this block
		for _, signature := range bftMeta.Signatures {
			if signature.SignerId == consenter.ConsenterId {
				ordererSignatures = append(ordererSignatures, &block.OrdererSignature{
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

// Writes ONLY VALID writes from block
func Writes(b *block.Block) []*Write {
	var blockWrites []*Write

	for _, e := range b.ValidEnvelopes() {
		for _, a := range e.TxActions() {
			for _, rwSet := range a.NsReadWriteSet() {
				for _, write := range rwSet.GetRwset().GetWrites() {
					blockWrite := &Write{
						KWWrite: write,

						Block:            b.GetHeader().GetNumber(),
						Chaincode:        a.ChaincodeSpec().GetChaincodeId().GetName(),
						ChaincodeVersion: a.ChaincodeSpec().GetChaincodeId().GetVersion(),
						Tx:               e.ChannelHeader().GetTxId(),
						Timestamp:        e.ChannelHeader().GetTimestamp(),
					}

					blockWrite.KeyObjectType, blockWrite.KeyAttrs = SplitCompositeKey(write.Key)
					// Normalized key without null byte
					blockWrite.Key = strings.Join(append([]string{blockWrite.KeyObjectType}, blockWrite.KeyAttrs...), "_")

					blockWrites = append(blockWrites, blockWrite)
				}
			}
		}
	}

	return blockWrites
}
