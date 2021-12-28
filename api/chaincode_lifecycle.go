package api

import (
	"context"

	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

// Lifecycle contains methods for interacting with system _lifecycle chaincode
type Lifecycle interface {
	// QueryInstalledChaincodes returns installed chaincodes list
	QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error)

	// InstallChaincode install chaincode on a peer
	InstallChaincode(ctx context.Context, installArgs *lb.InstallChaincodeArgs) (*lb.InstallChaincodeResult, error)

	// ApproveFromMyOrg approves chaincode package on a channel
	ApproveFromMyOrg(ctx context.Context, channel Channel, approvseArgs *lb.ApproveChaincodeDefinitionForMyOrgArgs) error

	// CheckCommitReadiness returns commitments statuses of participants on chaincode definition
	CheckCommitReadiness(ctx context.Context, channelID string, args *lb.CheckCommitReadinessArgs) (*lb.CheckCommitReadinessResult, error)
}
