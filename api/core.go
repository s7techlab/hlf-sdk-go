package api

import (
	"context"

	"github.com/hyperledger/fabric/msp"
)

type Channel interface {
	// Chaincode returns chaincode instance by chaincode name
	Chaincode(name string) Chaincode
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
}

// SystemCC describes interface to access Fabric System Chaincodes
type SystemCC interface {
	CSCC() CSCC
	QSCC() QSCC
	LSCC() LSCC
	Lifecycle() Lifecycle
}
