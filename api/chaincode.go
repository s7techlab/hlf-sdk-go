package api

import (
	"github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
)

type ChaincodeTx string

// Chaincode describes common operations with chaincode
type Chaincode interface {
	// Invoke returns invoke builder for presented chaincode function
	Invoke(fn string) ChaincodeInvokeBuilder
	// Query returns query builder for presented function and arguments
	Query(fn string, args ...string) ChaincodeQueryBuilder
	// Install fetches chaincode from repository and installs it on local peer
	Install(version string)
	// Subscribe returns subscription on chaincode events
	Subscribe(seekOption ...EventCCSeekOption) EventCCSubscription
}

// ChaincodeInvokeBuilder describes possibilities how to get invoke results
type ChaincodeInvokeBuilder interface {
	// WithIdentity allows to invoke chaincode from custom identity
	WithIdentity(identity msp.SigningIdentity) ChaincodeInvokeBuilder
	// Async lets get result of invoke without waiting of block commit
	Async() ChaincodeInvokeBuilder
	// ArgBytes invokes chaincode with slice of bytes as argument
	ArgBytes([][]byte) (ChaincodeTx, []byte, error)
	// ArgJSON invokes chaincode with slice of JSON-marshalled data
	ArgJSON(in ...interface{}) (ChaincodeTx, []byte, error)
	// ArgString invokes chaincode with slice of strings as arguments
	ArgString(args ...string) (ChaincodeTx, []byte, error)
}

// ChaincodeQueryBuilder describe possibilities how to get query results
type ChaincodeQueryBuilder interface {
	// WithIdentity allows to query chaincode from custom identity
	WithIdentity(identity msp.SigningIdentity) ChaincodeQueryBuilder
	// AsBytes allows to get result of querying chaincode as byte slice
	AsBytes() ([]byte, error)
	// AsJSON allows to get result of querying chaincode to presented structures using JSON-unmarshalling
	AsJSON(out interface{}) error
}

// CSCC describes Configuration System Chaincode (CSCC)
type CSCC interface {
	// JoinChain allows to join channel using presented genesis block
	JoinChain(channelName string, genesisBlock *common.Block) error
	// GetConfigBlock returns genesis block of channel
	GetConfigBlock(channelName string) (*common.Block, error)
	// GetConfigTree returns configuration tree of channel
	GetConfigTree(channelName string) (*peer.ConfigTree, error)
	// Channels returns list of joined channels
	Channels() (*peer.ChannelQueryResponse, error)
}

// QSCC describes Query System Chaincode (QSCC)
type QSCC interface {
	// GetChainInfo allows to get common info about channel blockchain
	GetChainInfo(channelName string) (*common.BlockchainInfo, error)
	// GetBlockByNumber allows to get block by number
	GetBlockByNumber(channelName string, blockNumber int64) (*common.Block, error)
	// GetBlockByHash allows to get block by hash
	GetBlockByHash(channelName string, blockHash []byte) (*common.Block, error)
	// GetTransactionByID allows to get transaction by id
	GetTransactionByID(channelName string, tx ChaincodeTx) (*peer.ProcessedTransaction, error)
	// GetBlockByTxID allows to get block by transaction
	GetBlockByTxID(channelName string, tx ChaincodeTx) (*common.Block, error)
}

// LSCC describes Life Cycle System Chaincode (LSCC)
type LSCC interface {
	// GetChaincodeData returns information about instantiated chaincode on target channel
	GetChaincodeData(channelName string, ccName string) (*ccprovider.ChaincodeData, error)
	// GetInstalledChaincodes returns list of installed chaincodes on peer
	GetInstalledChaincodes() (*peer.ChaincodeQueryResponse, error)
	// GetChaincodes returns list of instantiated chaincodes on channel
	GetChaincodes(channelName string) (*peer.ChaincodeQueryResponse, error)
	// GetDeploymentSpec returns spec for installed chaincode
	GetDeploymentSpec(channelName string, ccName string) (*peer.ChaincodeDeploymentSpec, error)
	// Install allows to install chaincode using deployment specification
	Install(spec *peer.ChaincodeDeploymentSpec) error
}
