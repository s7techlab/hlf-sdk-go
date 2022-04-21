package system

import (
	"github.com/s7techlab/hlf-sdk-go/api"
)

type scc struct {
	core api.Core
}

var _ api.SystemCC = (*scc)(nil)

func (c *scc) Lifecycle() api.Lifecycle {
	return NewLifecycle(c.core)
}

func NewSCC(core api.Core) api.SystemCC {
	return &scc{core: core}
}
