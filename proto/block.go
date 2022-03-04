package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/util/txflags"
)

type Block struct {
	Header          *common.BlockHeader     `json:"header"`
	Envelopes       []*Envelope             `json:"envelopes"`
	OrdererIdentity *msp.SerializedIdentity `json:"orderer_identity"`
}

func ParseBlock(block *common.Block) (*Block, error) {
	var err error
	parsedBlock := &Block{
		Header: block.Header,
	}

	txFilter := txflags.ValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	if parsedBlock.Envelopes, err = ParseEnvelopes(block.GetData().GetData(), txFilter); err != nil {
		return nil, err
	}

	parsedBlock.OrdererIdentity, err = ParseOrdererIdentity(block)
	if err != nil {
		return nil, fmt.Errorf("parsing orderer identity from block: %w", err)
	}

	return parsedBlock, nil
}

func ParseOrdererIdentity(cb *common.Block) (*msp.SerializedIdentity, error) {
	meta, err := protoutil.GetMetadataFromBlock(cb, 0)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata from block#%v. err %w", cb.Header.Number, err)
	}

	// TODO check transaction type? what transactions **definitely** have signatures?
	// some transactions have no signatures
	if len(meta.Signatures) == 0 {
		return nil, nil
	}
	// first signature by orderer
	signatureHeader, err := protoutil.UnmarshalSignatureHeader(meta.Signatures[0].SignatureHeader)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling header signature in block #%v. err: %w", cb.Header.Number, err)
	}

	serializedIndentity := &msp.SerializedIdentity{}

	err = proto.Unmarshal(signatureHeader.Creator, serializedIndentity)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling serialized indentity in block #%v. err: %w", cb.Header.Number, err)
	}

	return serializedIndentity, nil
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
