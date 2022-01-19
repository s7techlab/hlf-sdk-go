package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

type Channel interface {
	// Chaincode returns chaincode instance by chaincode name
	Chaincode(ctx context.Context, name string) (Chaincode, error)
	// Join channel
	Join(ctx context.Context) error
}

type Core interface {
	// Channel returns channel instance by channel name
	Channel(name string) Channel
	// CurrentIdentity identity returns current signing identity used by core
	CurrentIdentity() msp.SigningIdentity
	// CryptoSuite returns current crypto suite implementation
	CryptoSuite() CryptoSuite
	// System allows access to system chaincodes
	System() SystemCC
	// Current peer pool
	PeerPool() PeerPool
	// Chaincode installation
	Chaincode(name string) ChaincodePackage
	// FabricV2 returns if core works in fabric v2 mode
	FabricV2() bool
	// Events - shortcut for PeerPool().DeliverClient(...).SubscribeCC(...).Events()
	// subscribe on chaincode events using name of channel, chaincode and block offset
	// if provided 'identity' is 'nil' default one will be set
	Events(
		ctx context.Context,
		channelName string,
		ccName string,
		identity msp.SigningIdentity,
		blockRange ...int64,
	) (chan *peer.ChaincodeEvent, error)
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

// types which identify tx "wait'er" policy
// we dont make it as alias for preventing binding to our lib
const (
	TxWaiterSelfType string = "self"
	TxWaiterAllType  string = "all"
)

// SystemCC describes interface to access Fabric System Chaincodes
type SystemCC interface {
	CSCC() CSCC
	QSCC() QSCC
	LSCC() LSCC
	Lifecycle() Lifecycle
}
