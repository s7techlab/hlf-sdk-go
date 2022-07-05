package proto

import (
	"fmt"

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

var configBlocks map[channelName]*common.Block

func init() {
	configBlocks = make(map[channelName]*common.Block)
}

type Block struct {
	Header            *common.BlockHeader       `json:"header"`
	Envelopes         []*Envelope               `json:"envelopes"`
	OrdererIdentities []*msp.SerializedIdentity `json:"orderer_identities"`
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
	if raftOrdererIdentity != nil {
		parsedBlock.OrdererIdentities = append(parsedBlock.OrdererIdentities, raftOrdererIdentity)
	}

	// parse BFT orderer identities
	if block.Header.Number == 0 {
		configBlocks[channelName(parsedBlock.Envelopes[0].ChannelHeader.ChannelId)] = block
	} else {
		var bftOrdererIdentities []*msp.SerializedIdentity
		bftOrdererIdentities, err = ParseBTFOrderersIdentities(block, configBlocks[channelName(parsedBlock.Envelopes[0].ChannelHeader.ChannelId)])
		if err != nil {
			return nil, fmt.Errorf("parsing bft orderers identities: %w", err)
		}

		parsedBlock.OrdererIdentities = append(parsedBlock.OrdererIdentities, bftOrdererIdentities...)
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
	md := new(bftcommon.BFTMetadata)
	if err := proto.Unmarshal(block.Metadata.Metadata[common.BlockMetadataIndex_SIGNATURES], md); err != nil {
		return nil, fmt.Errorf("unmarshaling bft block metadata from metadata: %w", err)
	}

	lc := common.LastConfig{}
	err := proto.Unmarshal(md.Value, &lc)
	if err != nil {
		return nil, nil
	}

	configEnvelope, err := createConfigEnvelope(configBlock.Data.Data[0])
	if err != nil {
		return nil, nil
	}

	consensusType := configEnvelope.Config.ChannelGroup.Groups["Orderer"].Values["ConsensusType"].Value

	ct := &orderer.ConsensusType{}
	err = proto.Unmarshal(consensusType, ct)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling consensus type from config envelope concensus type: %w", err)
	}

	m := &bft.ConfigMetadata{}
	err = proto.Unmarshal(ct.Metadata, m)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling bft config metadata from concensus type metadata: %w", err)
	}

	var identities []*msp.SerializedIdentity
	identity := msp.SerializedIdentity{}
	for _, consenter := range m.Consenters {
		if err = proto.Unmarshal(consenter.Identity, &identity); err != nil {
			return nil, err
		}

		// among all channel orderers, find those that signed this block
		for _, signature := range md.Signatures {
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
