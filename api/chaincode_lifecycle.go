package api

import (
	"context"

	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

// Lifecycle contains methods for interacting with system _lifecycle chaincode
type Lifecycle interface {
	// QueryInstalledChaincode returns chaincode package installed on peer
	QueryInstalledChaincode(ctx context.Context, args *lb.QueryInstalledChaincodeArgs) (
		*lb.QueryInstalledChaincodeResult, error)

	// QueryInstalledChaincodes returns chaincode packages list installed on peer
	QueryInstalledChaincodes(ctx context.Context) (*lb.QueryInstalledChaincodesResult, error)

	// InstallChaincode sets up chaincode package on peer
	InstallChaincode(ctx context.Context, args *lb.InstallChaincodeArgs) (*lb.InstallChaincodeResult, error)

	// ApproveChaincodeDefinitionForMyOrg marks chaincode definition on a channel
	ApproveChaincodeDefinitionForMyOrg(ctx context.Context, channel string, args *lb.ApproveChaincodeDefinitionForMyOrgArgs) error

	// QueryApprovedChaincodeDefinition returns approved chaincode definition
	QueryApprovedChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryApprovedChaincodeDefinitionArgs) (
		*lb.QueryApprovedChaincodeDefinitionResult, error)

	// CheckCommitReadiness returns commitments statuses of participants on chaincode definition
	CheckCommitReadiness(ctx context.Context, channel string, args *lb.CheckCommitReadinessArgs) (
		*lb.CheckCommitReadinessResult, error)

	// CommitChaincodeDefinition the chaincode definition on the channel
	CommitChaincodeDefinition(ctx context.Context, channel string, args *lb.CommitChaincodeDefinitionArgs) (
		*lb.CommitChaincodeDefinitionResult, error)

	// QueryChaincodeDefinition returns chaincode definition committed on the channel
	QueryChaincodeDefinition(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionArgs) (
		*lb.QueryChaincodeDefinitionResult, error)

	// QueryChaincodeDefinitions returns chaincode definitions committed on the channel
	QueryChaincodeDefinitions(ctx context.Context, channel string, args *lb.QueryChaincodeDefinitionsArgs) (
		*lb.QueryChaincodeDefinitionsResult, error)
}
