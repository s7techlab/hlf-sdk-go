package api

import (
	"context"
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

type ChaincodePackage interface {
	// Allows to get latest version of chaincode
	Latest(ctx context.Context) (*peer.ChaincodeDeploymentSpec, error)
	// Installs chaincode using defined chaincode fetcher
	Install(ctx context.Context, path, version string) error
	// Instantiate chaincode on channel with presented params
	Instantiate(ctx context.Context, channelName, path, version, policy string, args [][]byte, transArgs TransArgs) error
}

type ChaincodeInvokeResponse struct {
	TxID    ChaincodeTx
	Payload []byte
	Err     error
}

// TxWaiter is interface for build your custom function for wait of result of tx after endorsement
type TxWaiter interface {
	Wait(ctx context.Context, txid ChaincodeTx) error
}

type DoOptions struct {
	DiscoveryChaincode *DiscoveryChaincode
	Channel            string
	Identity           msp.SigningIdentity
	Pool               PeerPool

	TxWaiter TxWaiter
}

type DoOption func(opt *DoOptions) error

// ChaincodeInvokeBuilder describes possibilities how to get invoke results
type ChaincodeInvokeBuilder interface {
	// WithIdentity allows to invoke chaincode from custom identity
	WithIdentity(identity msp.SigningIdentity) ChaincodeInvokeBuilder
	// Transient allows to pass arguments to transient map
	Transient(args TransArgs) ChaincodeInvokeBuilder
	// ArgBytes set slice of bytes as argument
	ArgBytes([][]byte) ChaincodeInvokeBuilder
	// ArgJSON set slice of JSON-marshalled data
	ArgJSON(in ...interface{}) ChaincodeInvokeBuilder
	// ArgString set slice of strings as arguments
	ArgString(args ...string) ChaincodeInvokeBuilder
	// Do makes invoke with built arguments
	Do(ctx context.Context, opts ...DoOption) (*peer.Response, ChaincodeTx, error)
}

// ChaincodeQueryBuilder describe possibilities how to get query results
type ChaincodeQueryBuilder interface {
	// WithIdentity allows to invoke chaincode from custom identity
	WithIdentity(identity msp.SigningIdentity) ChaincodeQueryBuilder
	// Transient allows to pass arguments to transient map
	Transient(args TransArgs) ChaincodeQueryBuilder
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

type CCFetcher interface {
	Fetch(ctx context.Context, id *peer.ChaincodeID) (*peer.ChaincodeDeploymentSpec, error)
}
