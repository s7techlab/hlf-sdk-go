package peer

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	fabricPeer "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/util"
)

var (
	errEndorsersEmpty = errors.New(`no endorsers`)
)

type processor struct {
	channelName string
}

func (p *processor) CreateProposal(cc *api.DiscoveryChaincode, identity msp.SigningIdentity, fn string, args [][]byte) (*fabricPeer.SignedProposal, api.ChaincodeTx, error) {
	invSpec, err := p.invocationSpec(cc, fn, args)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get invocation spec`)
	}

	extension := &fabricPeer.ChaincodeHeaderExtension{ChaincodeId: &fabricPeer.ChaincodeID{Name: cc.Name}}

	txId, nonce, err := util.NewTxWithNonce(identity)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get tx id`)
	}

	chHeader, err := util.NewChannelHeader(common.HeaderType_ENDORSER_TRANSACTION, txId, p.channelName, 0, extension)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get channel header`)
	}

	// TODO allow to pass TransientMap
	proposalPayload, err := proto.Marshal(&fabricPeer.ChaincodeProposalPayload{Input: invSpec, TransientMap: nil})
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to marshal proposal payload`)
	}

	sigHeader, err := util.NewSignatureHeader(identity, nonce)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get signatire header`)
	}

	header, err := proto.Marshal(&common.Header{
		ChannelHeader:   chHeader,
		SignatureHeader: sigHeader,
	})
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to marshal transaction header`)
	}

	proposal, err := proto.Marshal(&fabricPeer.Proposal{
		Header:  header,
		Payload: proposalPayload,
	})
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to marshal proposal`)
	}

	signedBytes, err := identity.Sign(proposal)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to sign proposal bytes`)
	}

	return &fabricPeer.SignedProposal{ProposalBytes: proposal, Signature: signedBytes}, api.ChaincodeTx(txId), nil
}

func (*processor) Send(proposal *fabricPeer.SignedProposal, peers ...api.Peer) ([]*fabricPeer.ProposalResponse, error) {
	if len(peers) == 0 {
		return nil, errEndorsersEmpty
	}

	respList := make([]*fabricPeer.ProposalResponse, 0)

	for _, p := range peers {
		if resp, err := p.Endorse(proposal); err != nil {
			return nil, errors.Wrap(err, `failed to collect response from `+p.Uri())
		} else {
			respList = append(respList, resp)
		}
	}
	return respList, nil
}

func (p *processor) invocationSpec(ccDef *api.DiscoveryChaincode, fn string, args [][]byte) ([]byte, error) {
	spec := &fabricPeer.ChaincodeInvocationSpec{
		ChaincodeSpec: &fabricPeer.ChaincodeSpec{
			Type:        ccDef.GetFabricType(),
			ChaincodeId: &fabricPeer.ChaincodeID{Name: ccDef.Name},
			Input:       &fabricPeer.ChaincodeInput{Args: p.prepareArgs(fn, args)},
		},
	}

	if specBytes, err := proto.Marshal(spec); err != nil {
		return nil, errors.Wrap(err, `failed to marshal spec to protobuf`)
	} else {
		return specBytes, nil
	}
}

// prepareArgs makes slice of strings to slice of slices of bytes
func (p *processor) prepareArgs(fn string, args [][]byte) [][]byte {
	byteArgs := make([][]byte, 0)
	byteArgs = append(byteArgs, []byte(fn))
	byteArgs = append(byteArgs, args...)
	return byteArgs
}

func NewProcessor(channelName string) api.PeerProcessor {
	return &processor{channelName: channelName}
}
