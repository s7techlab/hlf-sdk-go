package api

import (
	"context"

	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

// Lifecycle contains methods for interacting with system _lifecycle chaincode
type Lifecycle interface {
	// QueryInstalled returns chaincode packages list installed on peer
	QueryInstalled(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error)

	// Install sets up chaincode package on peer
	Install(ctx context.Context, args *lb.InstallChaincodeArgs) (*lb.InstallChaincodeResult, error)

	// Approve marks chaincode definition on a channel
	Approve(ctx context.Context, channel string, args *lb.ApproveChaincodeDefinitionForMyOrgArgs) error

	// QueryApproved returns approved chaincode definition
	QueryApproved(ctx context.Context, channel string, args *lb.QueryApprovedChaincodeDefinitionArgs) (*lb.QueryApprovedChaincodeDefinitionResult, error)

	// CheckReadiness returns commitments statuses of participants on chaincode definition
	CheckReadiness(ctx context.Context, channel string, args *lb.CheckCommitReadinessArgs) (*lb.CheckCommitReadinessResult, error)

	// Commit the chaincode definition on the channel
	Commit(ctx context.Context, channel string, args *lb.CommitChaincodeDefinitionArgs) (*lb.CommitChaincodeDefinitionResult, error)

	// QueryChaincodeDefinition returns chaincode definition committed on the channel
	QueryChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionArgs) (*lb.QueryChaincodeDefinitionResult, error)

	// QueryChaincodeDefinitions returns chaincode definitions committed on the channel
	QueryChaincodeDefinitions(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionsArgs) (*lb.QueryChaincodeDefinitionsResult, error)
}
