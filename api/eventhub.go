package api

import (
	"math"

	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

var (
	oldest  = &orderer.SeekPosition{Type: &orderer.SeekPosition_Oldest{Oldest: &orderer.SeekOldest{}}}
	newest  = &orderer.SeekPosition{Type: &orderer.SeekPosition_Newest{Newest: &orderer.SeekNewest{}}}
	maxStop = &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: math.MaxUint64}}}
)

type EventHub interface {
	// SubscribeCC allows to subscribe on chaincode events using name of channel, chaincode and block offset
	SubscribeCC(channelName string, ccName string, seekOpt ...EventCCSeekOption) EventCCSubscription
	// SubscribeTx allows to subscribe on transaction events by id
	// TODO (not implemented)
	SubscribeTx(channelName string, tx ChaincodeTx) EventTxSubscription
	// Close terminates eventHub grpc connection
	Close() error
}

type EventCCSeekOption func() (*orderer.SeekPosition, *orderer.SeekPosition)

// SeekNewest sets offset to new channel blocks
func SeekNewest() EventCCSeekOption {
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		return newest, maxStop
	}
}

// SeekOldest sets offset to channel blocks from beginning
func SeekOldest() EventCCSeekOption {
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		return oldest, maxStop
	}
}

// SeekSingle sets offset from block number
func SeekSingle(num uint64) EventCCSeekOption {
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		pos := &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: num}}}
		return pos, pos
	}
}

// SeekRange sets offset from one block to another by their numbers
func SeekRange(start, end uint64) EventCCSeekOption {
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		return &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: start}}},
			&orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: end}}}
	}
}

type EventCCSubscription interface {
	// Events initiates internal GRPC stream and returns channel on chaincode events or error if stream is failed
	Events() (chan *peer.ChaincodeEvent, error)
	// Close terminates internal GRPC stream
	Close() error
}

type EventTxSubscription interface {
	Result() (chan TxEvent, error)
	Close() error
}

type TxEvent struct {
	TxId    ChaincodeTx
	Success bool
	Error   error
}
