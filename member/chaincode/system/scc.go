package system

import (
	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/api"
)

const (
	csccName = `cscc`
	qsccName = `qscc`
	lsccName = `lscc`
)

type scc struct {
	peer     api.Peer
	identity msp.SigningIdentity
}

func (c *scc) QSCC() api.QSCC {
	return NewQSCC(c.peer, c.identity)
}

func (c *scc) CSCC() api.CSCC {
	return NewCSCC(c.peer, c.identity)
}

func (c *scc) LSCC() api.LSCC {
	return NewLSCC(c.peer, c.identity)
}

func NewSCC(peer api.Peer, identity msp.SigningIdentity) api.SystemCC {
	return &scc{peer: peer, identity: identity}
}
