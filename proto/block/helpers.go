package block

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/protobuf/encoding/protojson"

	hlfsdkgo "github.com/s7techlab/hlf-sdk-go"
)

func (x *Block) ValidEnvelopes() []*Envelope {
	var envs []*Envelope
	for _, e := range x.GetData().GetEnvelopes() {
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

func (x *Transaction) Events() []*peer.ChaincodeEvent {
	var events []*peer.ChaincodeEvent
	for _, a := range x.Actions {
		event := a.GetPayload().GetAction().GetProposalResponsePayload().GetExtension().GetEvents()
		if event != nil {
			events = append(events, event)
		}
	}
	return events
}

func (x *TransactionAction) Event() *peer.ChaincodeEvent {
	return x.GetPayload().GetAction().GetProposalResponsePayload().GetExtension().GetEvents()
}

func (x *TransactionAction) NsReadWriteSet() []*NsReadWriteSet {
	return x.GetPayload().GetAction().GetProposalResponsePayload().GetExtension().GetResults().GetNsRwset()
}

func (x *TransactionAction) ChaincodeSpec() *peer.ChaincodeSpec {
	return x.GetPayload().GetChaincodeProposalPayload().GetInput().GetChaincodeSpec()
}

func (x *TransactionAction) Endorsements() []*Endorsement {
	return x.GetPayload().GetAction().GetEndorsement()
}

func (x *TransactionAction) Response() *peer.Response {
	return x.GetPayload().GetAction().GetProposalResponsePayload().GetExtension().GetResponse()
}

func (x *Envelope) ChannelHeader() *common.ChannelHeader {
	return x.GetPayload().GetHeader().GetChannelHeader()
}

func (x *Envelope) SignatureHeader() *SignatureHeader {
	return x.GetPayload().GetHeader().GetSignatureHeader()
}

func (x *Envelope) TxActions() []*TransactionAction {
	return x.GetPayload().GetTransaction().GetActions()
}

func (x *ChannelConfig) ToJSON() ([]byte, error) {
	opt := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	return opt.Marshal(x)
}

// GetAllCertificates - returns all(root, intermediate, admins) certificates from all MSPs'
func (x *ChannelConfig) GetAllCertificates() ([]*Certificate, error) {
	var certs []*Certificate

	for mspID := range x.Applications {
		cs, err := x.Applications[mspID].Msp.GetAllCertificates()
		if err != nil {
			return nil, fmt.Errorf("get all msps certificates: %w", err)
		}
		certs = append(certs, cs...)
	}

	for mspID := range x.Orderers {
		cs, err := x.Orderers[mspID].Msp.GetAllCertificates()
		if err != nil {
			return nil, fmt.Errorf("get all orderers certificates: %w", err)
		}
		certs = append(certs, cs...)
	}

	return certs, nil
}

func (x *ChannelConfig) FabricVersion() hlfsdkgo.FabricVersion {
	if x.Capabilities != nil {
		_, isFabricV2 := x.Capabilities.Capabilities["V2_0"]
		if isFabricV2 {
			return hlfsdkgo.FabricV2
		}
		return hlfsdkgo.FabricV1
	}
	return hlfsdkgo.FabricVersionUndefined
}
