package api

import (
	"context"

	"github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
)

type ChaincodeTx string

type TransArgs map[string][]byte

// Chaincode describes common operations with chaincode
type Chaincode interface {
	// Invoke returns invoke builder for presented chaincode function
	Invoke(fn string) ChaincodeInvokeBuilder
	// Query returns query builder for presented function and arguments
	Query(fn string, args ...string) ChaincodeQueryBuilder
	// Install fetches chaincode from repository and installs it on local peer
	Install(version string)
	// Subscribe returns subscription on chaincode events
	Subscribe(ctx context.Context) (EventCCSubscription, error)
}

type ChaincodeInvokeResponse struct {
	TxID    ChaincodeTx
	Payload []byte
	Err     error
}

// ChaincodeBaseBuilder describes common operations available for invoke and query
type ChaincodeBaseBuilder interface {
	// WithIdentity allows to invoke chaincode from custom identity
	WithIdentity(identity msp.SigningIdentity) ChaincodeBaseBuilder
	// Transient allows to pass arguments to transient map
	Transient(args TransArgs) ChaincodeBaseBuilder
}

// ChaincodeInvokeBuilder describes possibilities how to get invoke results
type ChaincodeInvokeBuilder interface {
	ChaincodeBaseBuilder
	// Async lets get result of invoke without waiting of block commit
	Async(chan<- ChaincodeInvokeResponse) ChaincodeInvokeBuilder
	// ArgBytes set slice of bytes as argument
	ArgBytes([][]byte) ChaincodeInvokeBuilder
	// ArgJSON set slice of JSON-marshalled data
	ArgJSON(in ...interface{}) ChaincodeInvokeBuilder
	// ArgString set slice of strings as arguments
	ArgString(args ...string) ChaincodeInvokeBuilder
	// Do makes invoke with built arguments
	Do(ctx context.Context) (*peer.Response, ChaincodeTx, error)
}

// ChaincodeQueryBuilder describe possibilities how to get query results
type ChaincodeQueryBuilder interface {
	ChaincodeBaseBuilder
	// AsBytes allows to get result of querying chaincode as byte slice
	AsBytes(ctx context.Context) ([]byte, error)
	// AsJSON allows to get result of querying chaincode to presented structures using JSON-unmarshalling
	AsJSON(ctx context.Context, out interface{}) error
	// AsProposalResponse allows to get raw peer response
	AsProposalResponse(ctx context.Context) (*peer.ProposalResponse, error)
}

// CSCC describes Configuration System Chaincode (CSCC)
type CSCC interface {
	// JoinChain allows to join channel using presented genesis block
	JoinChain(ctx context.Context, channelName string, genesisBlock *common.Block) error
	// GetConfigBlock returns genesis block of channel
	GetConfigBlock(ctx context.Context, channelName string) (*common.Block, error)
	// GetConfigTree returns configuration tree of channel
	GetConfigTree(ctx context.Context, channelName string) (*peer.ConfigTree, error)
	// Channels returns list of joined channels
	Channels(ctx context.Context) (*peer.ChannelQueryResponse, error)
}

// QSCC describes Query System Chaincode (QSCC)
type QSCC interface {
	// GetChainInfo allows to get common info about channel blockchain
	GetChainInfo(ctx context.Context, channelName string) (*common.BlockchainInfo, error)
	// GetBlockByNumber allows to get block by number
	GetBlockByNumber(ctx context.Context, channelName string, blockNumber int64) (*common.Block, error)
	// GetBlockByHash allows to get block by hash
	GetBlockByHash(ctx context.Context, channelName string, blockHash []byte) (*common.Block, error)
	// GetTransactionByID allows to get transaction by id
	GetTransactionByID(ctx context.Context, channelName string, tx ChaincodeTx) (*peer.ProcessedTransaction, error)
	// GetBlockByTxID allows to get block by transaction
	GetBlockByTxID(ctx context.Context, channelName string, tx ChaincodeTx) (*common.Block, error)
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
	// Install allows to install chaincode using deployment specification
	Install(ctx context.Context, spec *peer.ChaincodeDeploymentSpec) error
}

type CCFetcher interface {
	Fetch(ctx context.Context, id *peer.ChaincodeID) (*peer.ChaincodeDeploymentSpec, error)
}
