package proto

import (
	"strings"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/peer"
)

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

// Writes ONLY VALID writes from block
func (x *Block) Writes() []*Write {
	var blockWrites []*Write

	for _, e := range x.ValidEnvelopes() {
		for _, a := range e.TxActions() {
			for _, rwSet := range a.NsReadWriteSet() {
				for _, write := range rwSet.Rwset.Writes {
					blockWrite := &Write{
						KWWrite: write,

						Block:            x.GetHeader().GetNumber(),
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
