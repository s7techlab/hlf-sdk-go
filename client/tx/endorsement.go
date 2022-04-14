package tx

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/proto"
)

var (
	ErrSignerNotDefined    = errors.New(`signer not defined`)
	ErrChaincodeNotDefined = errors.New(`chaincode not defined`)
)

type Endorsement struct {
	Channel      string
	Chaincode    string
	Args         [][]byte
	Signer       msp.SigningIdentity
	TransientMap map[string][]byte
}

func (e Endorsement) SignedProposal() (signedProposal *peer.SignedProposal, txID string, err error) {
	return NewEndorsementSignedProposal(e.Channel, e.Chaincode, e.Args, e.Signer, e.TransientMap)
}

func NewEndorsementSignedProposal(
	channel, chaincode string, args [][]byte, signer msp.SigningIdentity, transientMap map[string][]byte) (
	signedProposal *peer.SignedProposal, txID string, err error) {

	if chaincode == `` {
		return nil, ``, ErrChaincodeNotDefined
	}
	if signer == nil {
		return nil, ``, ErrSignerNotDefined
	}

	signerSerialized, err := signer.Serialize()
	if err != nil {
		return nil, ``, fmt.Errorf(`serialize signer: %w`, err)
	}

	txParams, err := GenerateParamsForSerializedIdentity(signerSerialized)
	if err != nil {
		return nil, ``, fmt.Errorf(`tx id: %w`, err)
	}

	header, err := proto.NewMarshalledCommonHeader(
		common.HeaderType_ENDORSER_TRANSACTION,
		txParams.ID,
		txParams.Nonce,
		txParams.Timestamp,
		signerSerialized,
		channel,
		chaincode)
	if err != nil {
		return nil, ``, fmt.Errorf(`tx header: %w`, err)
	}

	proposal, err := proto.NewMarshaledPeerProposal(header, chaincode, args, transientMap)
	if err != nil {
		return nil, ``, fmt.Errorf(`proposal: %w`, err)
	}

	signedProposal, err = proto.NewPeerSignedProposal(proposal, signer)

	return signedProposal, txID, err

}

func FnArgs(fn string, args [][]byte) [][]byte {
	return append([][]byte{[]byte(fn)}, args...)
}
