package peer

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/util"
)

type processor struct {
	channelName string
}

type endorseChannelResponse struct {
	Response *fabricPeer.ProposalResponse
	Error    error
}

func (p *processor) CreateProposal(chaincodeName string, identity msp.SigningIdentity, fn string, args [][]byte, transArgs api.TransArgs) (*fabricPeer.SignedProposal, api.ChaincodeTx, error) {
	invSpec, err := p.invocationSpec(chaincodeName, fn, args)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get invocation spec`)
	}

	extension := &fabricPeer.ChaincodeHeaderExtension{ChaincodeId: &fabricPeer.ChaincodeID{Name: chaincodeName}}

	txId, nonce, err := util.NewTxWithNonce(identity)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get tx id`)
	}

	chHeader, err := util.NewChannelHeader(common.HeaderType_ENDORSER_TRANSACTION, txId, p.channelName, 0, extension)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get channel header`)
	}

	proposalPayload, err := proto.Marshal(&fabricPeer.ChaincodeProposalPayload{Input: invSpec, TransientMap: transArgs})
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to marshal proposal payload`)
	}

	sigHeader, err := util.NewSignatureHeader(identity, nonce)
	if err != nil {
		return nil, ``, errors.Wrap(err, `failed to get signature header`)
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

func (*processor) Send(ctx context.Context, proposal *fabricPeer.SignedProposal, endorsingMspIDs []string, pool api.PeerPool) ([]*fabricPeer.ProposalResponse, error) {
	respList := make([]*fabricPeer.ProposalResponse, 0)
	respChan := make(chan endorseChannelResponse)

	// send all proposals concurrently
	for i := 0; i < len(endorsingMspIDs); i++ {
		go func(mspId string) {
			resp, err := pool.Process(ctx, mspId, proposal)
			respChan <- endorseChannelResponse{Response: resp, Error: err}
		}(endorsingMspIDs[i])
	}

	var errOccurred bool

	mErr := new(api.MultiError)

	// collecting peer responses
	for i := 0; i < len(endorsingMspIDs); i++ {
		resp := <-respChan
		if resp.Error != nil {
			errOccurred = true
			mErr.Add(resp.Error)
		}
		respList = append(respList, resp.Response)
	}

	if errOccurred {
		return respList, mErr
	}

	return respList, nil
}

func (p *processor) invocationSpec(chaincodeName string, fn string, args [][]byte) ([]byte, error) {
	spec := &fabricPeer.ChaincodeInvocationSpec{
		ChaincodeSpec: &fabricPeer.ChaincodeSpec{
			ChaincodeId: &fabricPeer.ChaincodeID{Name: chaincodeName},
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
