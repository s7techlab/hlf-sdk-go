package system

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/util"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	lsccPkg "github.com/hyperledger/fabric/core/scc/lscc"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/v2/peer"
)

type lscc struct {
	peerPool api.PeerPool
	identity msp.SigningIdentity
}

func (c *lscc) GetChaincodeData(ctx context.Context, channelName string, ccName string) (*ccprovider.ChaincodeData, error) {
	ccData := new(ccprovider.ChaincodeData)
	if resp, err := c.endorse(ctx, channelName, lsccPkg.GETCCDATA, util.ToChaincodeArgs(channelName, ccName), nil); err != nil {
		return nil, errors.Wrap(err, `failed to get chaincode data`)
	} else {
		return ccData, c.processEndorsementResponse(ccData, resp)
	}
}

func (c *lscc) GetInstalledChaincodes(ctx context.Context) (*peer.ChaincodeQueryResponse, error) {
	ccData := new(peer.ChaincodeQueryResponse)
	if resp, err := c.endorse(ctx, ``, lsccPkg.GETINSTALLEDCHAINCODES, nil, nil); err != nil {
		return nil, errors.Wrap(err, `failed to get chaincodes data`)
	} else {
		return ccData, c.processEndorsementResponse(ccData, resp)
	}
}

func (c *lscc) GetChaincodes(ctx context.Context, channelName string) (*peer.ChaincodeQueryResponse, error) {
	ccData := new(peer.ChaincodeQueryResponse)
	if resp, err := c.endorse(ctx, channelName, lsccPkg.GETCHAINCODES, nil, nil); err != nil {
		return nil, errors.Wrap(err, `failed to get chaincodes data`)
	} else {
		return ccData, c.processEndorsementResponse(ccData, resp)
	}
}

func (c *lscc) GetDeploymentSpec(ctx context.Context, channelName string, ccName string) (*peer.ChaincodeDeploymentSpec, error) {
	depData := new(peer.ChaincodeDeploymentSpec)
	if resp, err := c.endorse(ctx, channelName, lsccPkg.GETDEPSPEC, util.ToChaincodeArgs(channelName, ccName), nil); err != nil {
		return nil, errors.Wrap(err, `failed to get deployment data`)
	} else {
		return depData, c.processEndorsementResponse(depData, resp)
	}
}

func (c *lscc) Install(ctx context.Context, spec *peer.ChaincodeDeploymentSpec) error {
	if specBytes, err := proto.Marshal(spec); err != nil {
		return errors.Wrap(err, `failed to marshal protobuf`)
	} else {
		_, err = c.endorse(ctx, ``, lsccPkg.INSTALL, [][]byte{specBytes}, nil)
		return err
	}
}

func (c *lscc) Deploy(ctx context.Context, channelName string, spec *peer.ChaincodeDeploymentSpec, policy *common.SignaturePolicyEnvelope, opts ...api.LSCCDeployOption) (*peer.SignedProposal, *peer.ProposalResponse, error) {
	var (
		deployOpts api.LSCCDeployOptions
		err        error
	)

	for _, opt := range opts {
		if err = opt(&deployOpts); err != nil {
			return nil, nil, errors.Wrap(err, `failed to apply DeployOption`)
		}
	}

	args := make([][]byte, 0)
	// Append channel name to arguments
	args = append(args, []byte(channelName))
	// Append deployment spec to arguments
	specBytes, err := proto.Marshal(spec)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to marshal spec`)
	}
	args = append(args, specBytes)
	// Append endorsement policy to arguments
	policyBytes, err := proto.Marshal(policy)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to marshal endorsement policy`)
	}
	args = append(args, policyBytes)
	// Append escc name if presented or set default
	if deployOpts.Escc != `` {
		args = append(args, []byte(deployOpts.Escc))
	} else {
		args = append(args, []byte(`escc`))
	}
	// Append vscc name if presented or set default
	if deployOpts.Vscc != `` {
		args = append(args, []byte(deployOpts.Vscc))
	} else {
		args = append(args, []byte(`vscc`))
	}
	// Append private data collection config if presented
	if deployOpts.CollectionConfig != nil {
		ccConfigBytes, err := proto.Marshal(deployOpts.CollectionConfig)
		if err != nil {
			return nil, nil, errors.Wrap(err, `failed to marshal pvt data collection config`)
		}
		args = append(args, ccConfigBytes)
	}

	// Find chaincode instantiated or not
	ccList, err := c.GetChaincodes(ctx, channelName)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to fetch chaincode list`)
	}

	lsccCmd := lsccPkg.DEPLOY

	for _, cc := range ccList.Chaincodes {
		if cc.Name == spec.ChaincodeSpec.ChaincodeId.Name {
			lsccCmd = lsccPkg.UPGRADE
			break
		}
	}

	processor := peerSDK.NewProcessor(channelName)
	prop, _, err := processor.CreateProposal(lsccName, c.identity, lsccCmd, args, deployOpts.TransArgs)
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.endorseProposal(ctx, prop)
	return prop, resp, err
}

func (c *lscc) endorse(ctx context.Context, channelName string, fn string, args [][]byte, transArgs api.TransArgs) (*peer.ProposalResponse, error) {
	processor := peerSDK.NewProcessor(channelName)
	prop, _, err := processor.CreateProposal(lsccName, c.identity, fn, args, transArgs)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create proposal`)
	}

	return c.endorseProposal(ctx, prop)
}

func (c *lscc) endorseProposal(ctx context.Context, prop *peer.SignedProposal) (*peer.ProposalResponse, error) {
	resp, err := c.peerPool.Process(ctx, c.identity.GetMSPIdentifier(), prop)
	if err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
	}
	return resp, nil
}

func (c *lscc) processEndorsementResponse(out proto.Message, response *peer.ProposalResponse) error {
	if err := proto.Unmarshal(response.Response.Payload, out); err != nil {
		return errors.Wrap(err, `failed to unmarshal protobuf payload`)
	}

	return nil
}

func NewLSCC(peerPool api.PeerPool, identity msp.SigningIdentity) api.LSCC {
	return &lscc{peerPool: peerPool, identity: identity}
}
