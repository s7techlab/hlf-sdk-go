package api

import (
	"context"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

type EventsDeliverer interface {
	// Events - shortcut for PeerPool().DeliverClient(...).SubscribeCC(...).Events()
	// subscribe on chaincode events using name of channel, chaincode and block offset
	// if provided 'identity' is 'nil' default one will be set
	Events(
		ctx context.Context,
		channelName string,
		ccName string,
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
		channelName string,
		identity msp.SigningIdentity,
		blockRange ...int64,
	) (blockChan <-chan *common.Block, closer func() error, err error)
}
type Public interface {
	EventsDeliverer
	BlocksDeliverer

	// Invoke - shortcut for invoking chanincodes
	// if provided 'identity' is 'nil' default one will be set
	// txWaiterType - param which identify transaction waiting policy.
	// available: 'self'(wait for one peer of endorser org), 'all'(wait for each organizations from endorsement policy)
	// default is 'self'(even if you pass empty string)
	Invoke(
		ctx context.Context,
		chanName string,
		ccName string,
		args [][]byte,
		identity msp.SigningIdentity,
		transient map[string][]byte,
		txWaiterType string,
	) (res *peer.Response, chaincodeTx string, err error)

	// Query - shortcut for querying chanincodes
	// if provided 'identity' is 'nil' default one will be set
	Query(
		ctx context.Context,
		chanName string,
		ccName string,
		args [][]byte,
		identity msp.SigningIdentity,
		transient map[string][]byte,
	) (*peer.Response, error)
}
