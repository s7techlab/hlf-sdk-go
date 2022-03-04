package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/common/ccprovider"
)

type LSCCDeployOptions struct {
	Escc             string
	Vscc             string
	CollectionConfig *common.CollectionConfigPackage
	TransArgs        TransArgs
}

type LSCCDeployOption func(opts *LSCCDeployOptions) error

func WithCollectionConfig(config *common.CollectionConfigPackage) LSCCDeployOption {
	return func(opts *LSCCDeployOptions) error {
		opts.CollectionConfig = config
		return nil
	}
}

func WithESCC(escc string) LSCCDeployOption {
	return func(opts *LSCCDeployOptions) error {
		opts.Escc = escc
		return nil
	}
}

func WithVSCC(vscc string) LSCCDeployOption {
	return func(opts *LSCCDeployOptions) error {
		opts.Vscc = vscc
		return nil
	}
}

func WithTransientMap(args TransArgs) LSCCDeployOption {
	return func(opts *LSCCDeployOptions) error {
		opts.TransArgs = args
		return nil
	}
}

// LSCC describes Life Cycle System Chaincode (LSCC)
type LSCC interface {
	// GetChaincodeData returns information about instantiated chaincode on target channel
	GetChaincodeData(ctx context.Context, channelName string, ccName string) (*ccprovider.ChaincodeData, error)
	// GetInstalledChaincodes returns list of installed chaincodes on peer
	GetInstalledChaincodes(ctx context.Context) (*peer.ChaincodeQueryResponse, error)
	// GetChaincodes returns list of instantiated chaincodes on channel
	GetChaincodes(ctx context.Context, channelName string) (*peer.ChaincodeQueryResponse, error)
	// GetDeploymentSpec returns spec for installed chaincode
	GetDeploymentSpec(ctx context.Context, channelName string, ccName string) (*peer.ChaincodeDeploymentSpec, error)
	// Install allows installing chaincode using deployment specification
	Install(ctx context.Context, spec *peer.ChaincodeDeploymentSpec) error
	// Deploy allows instantiating or upgrade chaincode if instantiated
	// Currently, deploy method is not canonical as lscc implementation, but currently we need to get full proposal, and it's response to broadcast to orderer
	Deploy(ctx context.Context, channelName string, spec *peer.ChaincodeDeploymentSpec, policy *common.SignaturePolicyEnvelope, opts ...LSCCDeployOption) (*peer.SignedProposal, *peer.ProposalResponse, error)
}
