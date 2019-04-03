package chaincode

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/hyperledger/fabric/protos/peer"
)

type corePackage struct {
	ccName   string
	lscc     api.LSCC
	fetcher  api.CCFetcher
	orderer  api.Orderer
	identity msp.SigningIdentity
}

func (c *corePackage) Latest(ctx context.Context) (*peer.ChaincodeDeploymentSpec, error) {
	panic("implement me")
}

func (c *corePackage) Install(ctx context.Context, path, version string) error {
	depSpec, err := c.fetcher.Fetch(ctx, &peer.ChaincodeID{
		Name:    c.ccName,
		Path:    path,
		Version: version,
	})

	if err != nil {
		return errors.Wrap(err, `failed to fetch package`)
	}

	if err = c.lscc.Install(ctx, depSpec); err != nil {
		return errors.Wrap(err, `failed to invoke lifecycle chaincode`)
	}

	return nil
}

func (c *corePackage) Instantiate(ctx context.Context, channelName, path, version, policy string, args [][]byte, transArgs api.TransArgs) error {
	ePolicy, err := cauthdsl.FromString(policy)
	if err != nil {
		return errors.Wrap(err, `failed to parse endorsement policy`)
	}

	depSpec, err := c.fetcher.Fetch(ctx, &peer.ChaincodeID{
		Name:    c.ccName,
		Path:    path,
		Version: version,
	})

	depSpec.ChaincodeSpec.Input = &peer.ChaincodeInput{
		Args: args,
	}

	if err != nil {
		return errors.Wrap(err, `failed to fetch package`)
	}

	prop, resp, err := c.lscc.Deploy(ctx, channelName, depSpec, ePolicy, api.WithTransientMap(transArgs))
	if err != nil {
		return errors.Wrap(err, `failed to deploy chaincode`)
	}

	peerProp := new(peer.Proposal)
	err = proto.Unmarshal(prop.ProposalBytes, peerProp)
	if err != nil {
		return errors.Wrap(err, `failed to pnmarshal proposal for make peer.Proposal`)
	}

	env, err := utils.CreateSignedTx(peerProp, c.identity, resp)
	if err != nil {
		return errors.Wrap(err, "could not assemble transaction")
	}

	_, err = c.orderer.Broadcast(ctx, env)

	return err
}

func NewCorePackage(ccName string, lscc api.LSCC, fetcher api.CCFetcher, orderer api.Orderer, identity msp.SigningIdentity) api.ChaincodePackage {
	return &corePackage{ccName: ccName, lscc: lscc, fetcher: fetcher, orderer: orderer, identity: identity}
}
