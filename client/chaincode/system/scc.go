package system

import (
	"github.com/s7techlab/hlf-sdk-go/api"
)

const (
	csccName      = `cscc`
	qsccName      = `qscc`
	lsccName      = `lscc`
	lifecycleName = `_lifecycle`
)

type scc struct {
	core api.Core
}

var _ api.SystemCC = (*scc)(nil)

func (c *scc) QSCC() api.QSCC {
	return NewQSCC(c.core.PeerPool(), c.core.CurrentIdentity())
}

func (c *scc) CSCC() api.CSCC {
	if c.core.FabricV2() {
		return NewCSCCV2(c.core.PeerPool(), c.core.CurrentIdentity())
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
