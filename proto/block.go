package proto

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	bft "github.com/atomyze-ru/hlf-sdk-go/proto/smartbft"
	bftcommon "github.com/atomyze-ru/hlf-sdk-go/proto/smartbft/common"
	"github.com/atomyze-ru/hlf-sdk-go/util/txflags"
)

type channelName string

var (
	configBlocks map[channelName]*common.Block
	mu           sync.Mutex
)

func init() {
	mu.Lock()
	configBlocks = make(map[channelName]*common.Block)
	mu.Unlock()
}

type Block struct {
	Header            *common.BlockHeader       `json:"header"`
	Envelopes         []*Envelope               `json:"envelopes"`
	OrdererIdentities []*msp.SerializedIdentity `json:"orderer_identities"`
	BFTSignatures     [][]byte                  `json:"bft_signatures"`
}

func ParseBlock(block *common.Block) (*Block, error) {
	var err error
	parsedBlock := &Block{
		Header: block.Header,
	}

	txFilter := txflags.ValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	if parsedBlock.Envelopes, err = ParseEnvelopes(block.GetData().GetData(), txFilter); err != nil {
		return nil, fmt.Errorf("parsing envelopes: %w", err)
	}

	// parse Raft orderer identity
	raftOrdererIdentity, err := ParseOrdererIdentity(block)
	if err != nil {
		return nil, fmt.Errorf("parsing orderer identity from block: %w", err)
	}
	if raftOrdererIdentity != nil && raftOrdererIdentity.IdBytes != nil {
		parsedBlock.OrdererIdentities = append(parsedBlock.OrdererIdentities, raftOrdererIdentity)
	}

	// parse BFT orderer identities
	if block.Header.Number == 0 {
		mu.Lock()
		configBlocks[channelName(parsedBlock.Envelopes[0].ChannelHeader.ChannelId)] = block
		mu.Unlock()
	} else {
		mu.Lock()
		configBlock := configBlocks[channelName(parsedBlock.Envelopes[0].ChannelHeader.ChannelId)]
		mu.Unlock()

		var bftOrdererIdentities []*msp.SerializedIdentity
		bftOrdererIdentities, err = ParseBTFOrderersIdentities(block, configBlock)
		if err != nil {
			return nil, fmt.Errorf("parsing bft orderers identities: %w", err)
		}

		parsedBlock.OrdererIdentities = append(parsedBlock.OrdererIdentities, bftOrdererIdentities...)
		if len(bftOrdererIdentities) > 0 {
			var meta *common.Metadata
			meta, err = protoutil.GetMetadataFromBlock(block, common.BlockMetadataIndex_SIGNATURES)
			if err != nil {
				return nil, fmt.Errorf("get metadata from block: %w", err)
			}

			for _, signature := range meta.Signatures {
				parsedBlock.BFTSignatures = append(parsedBlock.BFTSignatures, signature.Signature)
			}
		}
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

	serializedIdentity := &msp.SerializedIdentity{}
	err = proto.Unmarshal(signatureHeader.Creator, serializedIdentity)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling serialized indentity from signature header: %w", err)
	}

	return serializedIdentity, nil
}

func ParseBTFOrderersIdentities(block *common.Block, configBlock *common.Block) ([]*msp.SerializedIdentity, error) {
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

	var identities []*msp.SerializedIdentity
	for _, consenter := range configMetadata.Consenters {
		var identity msp.SerializedIdentity
		if err = proto.Unmarshal(consenter.Identity, &identity); err != nil {
			return nil, err
		}

		// among all channel orderers, find those that signed this block
		for _, signature := range bftMeta.Signatures {
			if signature.SignerId == consenter.ConsenterId {
				identities = append(identities, &identity)
			}
		}
	}

	return identities, nil
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

func (b *Block) ValidEnvelopes() Envelopes {
	var envs Envelopes
	for _, e := range b.Envelopes {
		if e.ValidationCode != peer.TxValidationCode_VALID {
			continue
		}

		envs = append(envs, e)
	}

	return envs
}
