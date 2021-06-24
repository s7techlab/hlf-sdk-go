package system

import (
	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/api"
)

const (
	csccName      = `cscc`
	qsccName      = `qscc`
	lsccName      = `lscc`
	lifecycleName = `_lifecycle`
)

type scc struct {
	peerPool api.PeerPool
	identity msp.SigningIdentity
	fabricV2 bool
}

func (c *scc) QSCC() api.QSCC {
	return NewQSCC(c.peerPool, c.identity)
}

func (c *scc) CSCC() api.CSCC {
	if c.fabricV2 {
		return NewCSCCV2(c.peerPool, c.identity)
	}
	return NewCSCCV1(c.peerPool, c.identity)
}

func (c *scc) LSCC() api.LSCC {
	return NewLSCC(c.peerPool, c.identity)
}

func (c *scc) Lifecycle() api.Lifecycle {
	return NewLifecycle(c.peerPool, c.identity)
}

func NewSCC(peer api.PeerPool, identity msp.SigningIdentity, fabricV2 bool) api.SystemCC {
	return &scc{peerPool: peer, identity: identity, fabricV2: fabricV2}
}
