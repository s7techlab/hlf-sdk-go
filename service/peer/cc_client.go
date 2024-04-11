package peer

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
)

type (
	ChaincodeManagerClient interface {
		InstallChaincode(ctx context.Context, deploymentSpec *peer.ChaincodeDeploymentSpec) error
	}

	ChaincodeInfoClient interface {
		GetInstalledChaincodes(ctx context.Context) (*Chaincodes, error)
		GetInstantiatedChaincodes(ctx context.Context, channel string) (*Chaincodes, error)
	}

	LSCCChaincodeUpper interface {
		UpChaincode(ctx context.Context, depSpec *peer.ChaincodeDeploymentSpec, upChaincode *UpChaincodeRequest) (*UpChaincodeResponse, error)
	}

	LifecycleChaincodeUpper interface {
		UpChaincode(ctx context.Context, chaincode *Chaincode, upChaincode *UpChaincodeRequest) (*UpChaincodeResponse, error)
	}
)

//func LifecycleVersionMatch(version LifecycleVersion, fabricVersion chaincode.FabricVersion) bool {
//	switch version {
//	case LifecycleVersion_LIFECYCLE_V1:
//		switch fabricVersion {
//		case chaincode.FabricVersion_FABRIC_V2_LIFECYCLE:
//			return false
//		case chaincode.FabricVersion_FABRIC_V1:
//			fallthrough
//		case chaincode.FabricVersion_FABRIC_V2:
//			return true
//		}
//
//	case LifecycleVersion_LIFECYCLE_V2:
//		switch fabricVersion {
//
//		case chaincode.FabricVersion_FABRIC_V2_LIFECYCLE:
//			return true
//		case chaincode.FabricVersion_FABRIC_V1:
//			fallthrough
//		case chaincode.FabricVersion_FABRIC_V2:
//			return false
//		}
//	}
//
//	return false
//}
