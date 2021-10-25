package api

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
)

type Channel interface {
	// Chaincode returns chaincode instance by chaincode name
	Chaincode(ctx context.Context, name string) (Chaincode, error)
	// Joins channel
	Join(ctx context.Context) error
	// CSCC implements Configuration System Chaincode (CSCC)
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
	// ChannelChaincode - shortcut for Channel().Chaincode
	ChannelChaincode(ctx context.Context, chanName string, ccName string) (Chaincode, error)
	// Events - shortcut for PeerPool().DeliverClient(...).SubscribeCC(...).Events()
	// subscribe on chaincode events using name of channel, chaincode and block offset
	Events(
		ctx context.Context,
		channelName string,
		ccName string,
		eventCCSeekOption ...func() (*orderer.SeekPosition, *orderer.SeekPosition),
	) (chan *peer.ChaincodeEvent, error)
}

// SystemCC describes interface to access Fabric System Chaincodes
type SystemCC interface {
	CSCC() CSCC
	QSCC() QSCC
	LSCC() LSCC
	Lifecycle() Lifecycle
}
