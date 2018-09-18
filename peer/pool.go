package peer

import (
	"sync"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/logger"
	"go.uber.org/zap"
)

type peerPool struct {
	log *zap.Logger

	store   map[string]api.Peer
	storeMx sync.RWMutex
}

func (p *peerPool) Get(mspId string) (api.Peer, error) {
	p.storeMx.RLock()
	defer p.storeMx.RUnlock()
	if peer, ok := p.store[mspId]; ok {
		return peer, nil
	}

	return nil, api.ErrPeerPoolUnknownInstance
}

func (p *peerPool) Set(mspId string, peer api.Peer) error {
	p.storeMx.Lock()
	defer p.storeMx.Unlock()
	if _, ok := p.store[mspId]; ok {
		return api.ErrPeerAlreadySet
	} else {
		p.store[mspId] = peer
	}

	return nil
}

func NewPeerPool() api.PeerPool {
	return &peerPool{store: make(map[string]api.Peer), log: logger.DefaultLogger}
}
