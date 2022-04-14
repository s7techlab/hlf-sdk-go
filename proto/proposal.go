package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

func NewPeerSignedProposal(proposal []byte, identity msp.SigningIdentity) (*peer.SignedProposal, error) {
	signedProposal, err := identity.Sign(proposal)
	if err != nil {
		return nil, fmt.Errorf(`sign proposal: %w`, err)
	}

	return &peer.SignedProposal{
		ProposalBytes: proposal,
		Signature:     signedProposal}, nil
}

func NewMarshaledPeerProposal(header []byte, chaincode string, args [][]byte, transientMap map[string][]byte) ([]byte, error) {
	payload, err := NewMarshalledChaincodeProposalPayload(chaincode, args, transientMap)
	if err != nil {
		return nil, fmt.Errorf(`chaincode proposal payload: %w`, err)
	}

	return proto.Marshal(&peer.Proposal{
		Header:  header,
		Payload: payload,
	})
}

func NewMarshalledChaincodeProposalPayload(chaincode string, args [][]byte, transientMap map[string][]byte) ([]byte, error) {
	invSpec, err := NewMarshalledPeerChaincodeInvocationSpec(chaincode, args)
	if err != nil {
		return nil, fmt.Errorf(`invocation spec: %w`, err)
	}

	return proto.Marshal(&peer.ChaincodeProposalPayload{
		Input:        invSpec,
		TransientMap: transientMap,
	})
}

func NewMarshalledPeerChaincodeInvocationSpec(chaincode string, args [][]byte) ([]byte, error) {
	spec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{Name: chaincode},
			Input:       &peer.ChaincodeInput{Args: args},
		},
	}
	return proto.Marshal(spec)
}
