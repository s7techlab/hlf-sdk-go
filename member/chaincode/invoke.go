package chaincode

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	fabricPeer "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/peer"
)

type invokeBuilder struct {
	async     bool
	ccCore    *Core
	fn        string
	processor api.PeerProcessor
	identity  msp.SigningIdentity
}

func (b *invokeBuilder) WithIdentity(identity msp.SigningIdentity) api.ChaincodeInvokeBuilder {
	b.identity = identity
	return b
}

func (b *invokeBuilder) Async() api.ChaincodeInvokeBuilder {
	b.async = true
	return b
}

func (b *invokeBuilder) ArgBytes(args [][]byte) (api.ChaincodeTx, []byte, error) {
	endorsers, err := b.ccCore.dp.Endorsers(b.ccCore.channelName, b.ccCore.name)
	if err != nil {
		return ``, nil, errors.Wrap(err, `failed to get endorsers list`)
	}

	endorsersList := make([]api.Peer, 0)

	for _, ec := range endorsers {
		if p, err := peer.New(ec); err != nil {
			return ``, nil, errors.Wrap(err, `failed to initialize endorser`)
		} else {
			endorsersList = append(endorsersList, p)
		}
	}

	cc, err := b.ccCore.dp.Chaincode(b.ccCore.channelName, b.ccCore.name)
	if err != nil {
		return ``, nil, errors.Wrap(err, `failed to get chaincode definition`)
	}

	proposal, tx, err := b.processor.CreateProposal(cc, b.identity, b.fn, args)
	if err != nil {
		return tx, nil, errors.Wrap(err, `failed to get signed proposal`)
	}

	peerResponses, err := b.processor.Send(proposal, endorsersList...)
	if err != nil {
		return tx, nil, errors.Wrap(err, `failed to collect peer responses`)
	}

	envelope, err := b.getTransaction(proposal, peerResponses)
	if err != nil {
		return tx, nil, errors.Wrap(err, `failed to get envelope`)
	}

	_, err = b.ccCore.orderer.Broadcast(envelope)
	if err != nil {
		return tx, nil, errors.Wrap(err, `failed to get orderer response`)
	}

	if b.async {
		return tx, peerResponses[0].Response.Payload, nil
	} else {
		tsSub := b.ccCore.deliverClient.SubscribeTx(b.ccCore.channelName, tx)
		defer tsSub.Close()
		event, err := tsSub.Result()
		if err != nil {
			return tx, nil, errors.Wrap(err, `failed to subscribe on tx event`)
		} else {
			for ev := range event {
				if ev.Success {
					return tx, peerResponses[0].Response.Payload, nil
				} else {
					return tx, nil, errors.Wrap(ev.Error, `failed to get confirmation from endorser`)
				}
			}
			return tx, nil, errors.New(`failed to get tx event`)
		}
	}
}

func (b *invokeBuilder) getTransaction(proposal *fabricPeer.SignedProposal, peerResponses []*fabricPeer.ProposalResponse) (*common.Envelope, error) {

	endorsements := make([]*fabricPeer.Endorsement, 0)
	for _, resp := range peerResponses {
		endorsements = append(endorsements, resp.Endorsement)
	}

	prop := new(fabricPeer.Proposal)
	if err := proto.Unmarshal(proposal.ProposalBytes, prop); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal proposal protobuf`)
	}

	propHeader := new(common.Header)
	if err := proto.Unmarshal(prop.Header, propHeader); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal proposal header protobuf`)
	}

	propPayload := new(fabricPeer.ChaincodeProposalPayload)
	if err := proto.Unmarshal(prop.Payload, propPayload); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal proposal payload protobuf`)
	}

	ccProposalPayload, err := proto.Marshal(&fabricPeer.ChaincodeProposalPayload{
		Input:        propPayload.Input,
		TransientMap: propPayload.TransientMap,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to get chaincode proposal payload`)
	}

	ccActionPayload, err := proto.Marshal(&fabricPeer.ChaincodeActionPayload{
		Action: &fabricPeer.ChaincodeEndorsedAction{
			ProposalResponsePayload: peerResponses[0].Payload,
			Endorsements:            endorsements,
		},
		ChaincodeProposalPayload: ccProposalPayload,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to get chaincode action payload`)
	}

	txPayload, err := proto.Marshal(&fabricPeer.Transaction{
		Actions: []*fabricPeer.TransactionAction{{Header: propHeader.SignatureHeader, Payload: ccActionPayload}},
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to get transaction payload`)
	}

	commonPayload, err := proto.Marshal(&common.Payload{
		Header: propHeader,
		Data:   txPayload,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to get common payload`)
	}

	signedPayload, err := b.identity.Sign(commonPayload)
	if err != nil {
		return nil, errors.Wrap(err, `failed to sign common payload`)
	}

	return &common.Envelope{Payload: commonPayload, Signature: signedPayload}, nil

}

func (b *invokeBuilder) ArgJSON(in ...interface{}) (api.ChaincodeTx, []byte, error) {
	argBytes := make([][]byte, 0)
	for _, arg := range in {
		if data, err := json.Marshal(arg); err != nil {
			return ``, nil, errors.Wrap(err, `failed to marshal argument to JSON`)
		} else {
			argBytes = append(argBytes, data)
		}
	}
	return b.ArgBytes(argBytes)
}

func (b *invokeBuilder) ArgString(args ...string) (api.ChaincodeTx, []byte, error) {
	return b.ArgBytes(argsToBytes(args...))
}

func NewInvokeBuilder(ccCore *Core, fn string) api.ChaincodeInvokeBuilder {
	processor := peer.NewProcessor(ccCore.channelName)
	return &invokeBuilder{ccCore: ccCore, fn: fn, processor: processor, identity: ccCore.identity, async: false}
}
