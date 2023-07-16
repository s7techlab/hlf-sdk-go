package api

import (
	"context"

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
	// CurrentMspPeers returns current msp peers
	CurrentMspPeers() []Peer
	// CryptoSuite returns current crypto suite implementation
	CryptoSuite() CryptoSuite
	// PeerPool current peer pool
	PeerPool() PeerPool
	// FabricV2 returns if core works in fabric v2 mode
	FabricV2() bool

	Public
}

// types which identify tx "wait'er" policy
// we don't make it as alias for preventing binding to our lib
const (
	TxWaiterSelfType string = "self"
	TxWaiterAllType  string = "all"
)
