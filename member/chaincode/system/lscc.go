package system

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	lsccPkg "github.com/hyperledger/fabric/core/scc/lscc"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

type lscc struct {
	peer     api.Peer
	identity msp.SigningIdentity
}

func (c *lscc) GetChaincodeData(channelName string, ccName string) (*ccprovider.ChaincodeData, error) {
	ccData := new(ccprovider.ChaincodeData)
	if err := c.endorse(channelName, ccData, lsccPkg.GETCCDATA, channelName, ccName); err != nil {
		return nil, errors.Wrap(err, `failed to get chaincode data`)
	}

	return ccData, nil
}

func (c *lscc) GetInstalledChaincodes() (*peer.ChaincodeQueryResponse, error) {
	ccData := new(peer.ChaincodeQueryResponse)
	if err := c.endorse(``, ccData, lsccPkg.GETINSTALLEDCHAINCODES); err != nil {
		return nil, errors.Wrap(err, `failed to get chaincodes data`)
	}

	return ccData, nil
}

func (c *lscc) GetChaincodes(channelName string) (*peer.ChaincodeQueryResponse, error) {
	ccData := new(peer.ChaincodeQueryResponse)
	if err := c.endorse(channelName, ccData, lsccPkg.GETCHAINCODES); err != nil {
		return nil, errors.Wrap(err, `failed to get chaincodes data`)
	}

	return ccData, nil
}

func (c *lscc) GetDeploymentSpec(channelName string, ccName string) (*peer.ChaincodeDeploymentSpec, error) {
	depData := new(peer.ChaincodeDeploymentSpec)
	if err := c.endorse(channelName, depData, lsccPkg.GETDEPSPEC, channelName, ccName); err != nil {
		return nil, errors.Wrap(err, `failed to get deployment data`)
	}
	return depData, nil
}

func (c *lscc) Install(spec *peer.ChaincodeDeploymentSpec) error {
	if specBytes, err := proto.Marshal(spec); err != nil {
		return errors.Wrap(err, `failed to marshal protobuf`)
	} else {
		return c.endorse(``, nil, lsccPkg.INSTALL, string(specBytes))
	}
}

func (c *lscc) endorse(channelName string, out proto.Message, fn string, args ...string) error {
	processor := peerSDK.NewProcessor(channelName)
	prop, _, err := processor.CreateProposal(&api.DiscoveryChaincode{Name: lsccName, Type: api.CCTypeGoLang}, c.identity, fn, util.ToChaincodeArgs(args...))
	if err != nil {
		return errors.Wrap(err, `failed to create proposal`)
	}

	resp, err := c.peer.Endorse(prop)
	if err != nil {
		return errors.Wrap(err, `failed to endorse proposal`)
	}

	if out != nil {
		if err = proto.Unmarshal(resp.Response.Payload, out); err != nil {
			return errors.Wrap(err, `failed to unmarshal protobuf`)
		}
	}

	return nil
}

func NewLSCC(peer api.Peer, identity msp.SigningIdentity) api.LSCC {
	return &lscc{peer: peer, identity: identity}
}
