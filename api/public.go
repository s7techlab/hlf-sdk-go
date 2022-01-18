package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

type Public interface {

	// Events - shortcut for PeerPool().DeliverClient(...).SubscribeCC(...).Events()
	// subscribe on chaincode events using name of channel, chaincode and block offset
	// if provided 'identity' is 'nil' default one will be set
	Events(
		ctx context.Context,
		channelName string,
		ccName string,
		identity msp.SigningIdentity,
		blockRange ...int64,
	) (chan interface {
		Event() *peer.ChaincodeEvent
		Block() uint64
	}, error)

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
