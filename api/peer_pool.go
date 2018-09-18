package api

const (
	ErrPeerPoolUnknownInstance = Error(`unknown peer instance`)
	ErrPeerAlreadySet          = Error(`peer already set`)
)

type PeerPool interface {
	Get(mspId string) (Peer, error)
	Set(mspId string, api Peer) error
}
