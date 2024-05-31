package peer

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/s7techlab/hlf-sdk-go/proto/ccpackage"
	peerproto "github.com/s7techlab/hlf-sdk-go/proto/peer"
)

type (
	ChaincodeManagerClient interface {
		InstallChaincode(ctx context.Context, deploymentSpec *peer.ChaincodeDeploymentSpec) error
	}

	ChaincodeInfoClient interface {
		GetInstalledChaincodes(ctx context.Context) (*peerproto.Chaincodes, error)
		GetInstantiatedChaincodes(ctx context.Context, channel string) (*peerproto.Chaincodes, error)
	}

	LSCCChaincodeUpper interface {
		UpChaincode(ctx context.Context, depSpec *peer.ChaincodeDeploymentSpec, upChaincode *peerproto.UpChaincodeRequest) (
			*peerproto.UpChaincodeResponse, error)
	}

	LifecycleChaincodeUpper interface {
		UpChaincode(ctx context.Context, chaincode *peerproto.Chaincode, upChaincode *peerproto.UpChaincodeRequest) (
			*peerproto.UpChaincodeResponse, error)
	}
)

func LifecycleVersionMatch(version peerproto.LifecycleVersion, fabricVersion ccpackage.FabricVersion) bool {
	switch version {
	case peerproto.LifecycleVersion_LIFECYCLE_V1:
		switch fabricVersion {
		case ccpackage.FabricVersion_FABRIC_V2_LIFECYCLE:
			return false
		case ccpackage.FabricVersion_FABRIC_V1:
			fallthrough
		case ccpackage.FabricVersion_FABRIC_V2:
			return true
		}

	case peerproto.LifecycleVersion_LIFECYCLE_V2:
		switch fabricVersion {

		case ccpackage.FabricVersion_FABRIC_V2_LIFECYCLE:
			return true
		case ccpackage.FabricVersion_FABRIC_V1:
			fallthrough
		case ccpackage.FabricVersion_FABRIC_V2:
			return false
		}
	}

	return false
}
