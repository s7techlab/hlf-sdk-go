package api

import "github.com/pkg/errors"

var (
	ErrPeerPoolUnknownInstance = errors.New(`unknown peer instance`)
	ErrPeerAlreadySet          = errors.New(`peer already set`)
)

type PeerPool interface {
	Get(peerUri string) (Peer, error)
	Set(api Peer) error
}
