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

	// Commit the chaincode definition on the channel
	Commit(ctx context.Context, channel Channel, commitArgs *lb.CommitChaincodeDefinitionArgs) (*lb.CommitChaincodeDefinitionResult, error)

	// QueryChaincodeDefinition returns chaincode definition committed on the channel
	QueryChaincodeDefinition(ctx context.Context, channel Channel, args *lb.QueryChaincodeDefinitionArgs) (*lb.QueryChaincodeDefinitionResult, error)

	// QueryChaincodeDefinitions returns chaincode definitions committed on the channel
	QueryChaincodeDefinitions(ctx context.Context, channel Channel, args *lb.QueryChaincodeDefinitionsArgs) (*lb.QueryChaincodeDefinitionsResult, error)
}
