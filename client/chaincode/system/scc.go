package system

import (
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
)

const (
	csccName      = `cscc`
	qsccName      = `qscc`
	lsccName      = `lscc`
	lifecycleName = `_lifecycle`
)

type scc struct {
	core     api.Core
	peerPool api.PeerPool
	identity msp.SigningIdentity
	fabricV2 bool
}

func (c *scc) QSCC() api.QSCC {
	return NewQSCC(c.core.PeerPool(), c.core.CurrentIdentity())
}

func (c *scc) CSCC() api.CSCC {
	if c.fabricV2 {
		return NewCSCCV2(c.peerPool, c.identity)
	}
	return NewCSCCV1(c.core.PeerPool(), c.core.CurrentIdentity())
}

func (c *scc) LSCC() api.LSCC {
	return NewLSCC(c.core.PeerPool(), c.core.CurrentIdentity())
}

func (c *scc) Lifecycle() api.Lifecycle {
	return NewLifecycle(c.core)
}

func NewSCC(core api.Core) api.SystemCC {
	return &scc{core: core}
}
