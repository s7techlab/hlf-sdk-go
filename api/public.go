package api

import (
	"context"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

type CurrentIdentity interface {
	// CurrentIdentity identity returns current signing identity used by core
	CurrentIdentity() msp.SigningIdentity
}

type EventsDeliverer interface {
	// Events - shortcut for PeerPool().DeliverClient(...).SubscribeCC(...).Events()
	// subscribe on chaincode events using name of channel, chaincode and block offset
	// if provided 'identity' is 'nil' default one will be set
	Events(
		ctx context.Context,
		channel string,
		chaincode string,
		identity msp.SigningIdentity,
		blockRange ...int64,
	) (events chan interface {
		Event() *peer.ChaincodeEvent
		Block() uint64
		TxTimestamp() *timestamp.Timestamp
	}, closer func() error, err error)
}

type BlocksDeliverer interface {
	// Blocks - shortcut for core.PeerPool().DeliverClient(mspIdentity).SubscribeBlock(chanName,seekRange).Blocks()
	// subscribe to new blocks on specified channel
	// if provided 'identity' is 'nil' default one will be set
	Blocks(
		ctx context.Context,
		channel string,
		identity msp.SigningIdentity,
		blockRange ...int64,
	) (blockChan <-chan *common.Block, closer func() error, err error)
}

type Querier interface {
	CurrentIdentity
	// Query - shortcut for querying chaincodes
	// if provided 'identity' is 'nil' default one will be set
	Query(
		ctx context.Context,
		channel string,
		chaincode string,
		args [][]byte,
		identity msp.SigningIdentity,
		transient map[string][]byte,
	) (*peer.Response, error)
}

type Invoker interface {
	Querier

	// Invoke - shortcut for invoking chaincodes
	// if provided 'identity' is 'nil' default one will be set
	// txWaiterType - param which identify transaction waiting policy.
	// available: 'self'(wait for one peer of endorser org), 'all'(wait for each organization from endorsement policy)
	// default is 'self'(even if you pass empty string)
	Invoke(
		ctx context.Context,
		channel string,
		chaincode string,
		args [][]byte,
		identity msp.SigningIdentity,
		transient map[string][]byte,
		txWaiterType string,
	) (res *peer.Response, chaincodeTx string, err error)
}

type Public interface {
	EventsDeliverer
	BlocksDeliverer
	Invoker
}
