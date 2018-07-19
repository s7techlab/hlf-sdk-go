package member

import "github.com/s7techlab/hlf-sdk-go/api"

type coreOptions struct {
	peer    api.Peer
	orderer api.Orderer
}

// CoreOpt describes opt which will be applied to coreOptions
type CoreOpt func(c *coreOptions) error

// WithPeer allows to use custom instance of peer in core
func WithPeer(peer api.Peer) CoreOpt {
	return func(c *coreOptions) error {
		c.peer = peer
		return nil
	}
}

// WithOrderer allows to use custom instance of orderer in core
func WithOrderer(orderer api.Orderer) CoreOpt {
	return func(c *coreOptions) error {
		c.orderer = orderer
		return nil
	}
}
