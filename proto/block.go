package proto

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/s7techlab/hlf-sdk-go/v2/util/txflags"
)

type Block struct {
	Header    *common.BlockHeader
	Envelopes []*Envelope
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

	return parsedBlock, nil
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
