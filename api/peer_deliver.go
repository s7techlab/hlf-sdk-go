package api

import (
	"context"
	"math"

	"fmt"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

var (
	oldest  = &orderer.SeekPosition{Type: &orderer.SeekPosition_Oldest{Oldest: &orderer.SeekOldest{}}}
	newest  = &orderer.SeekPosition{Type: &orderer.SeekPosition_Newest{Newest: &orderer.SeekNewest{}}}
	maxStop = &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: math.MaxUint64}}}
)

type DeliverClient interface {
	// SubscribeCC allows to subscribe on chaincode events using name of channel, chaincode and block offset
	SubscribeCC(ctx context.Context, channelName string, ccName string, seekOpt ...EventCCSeekOption) EventCCSubscription
	// SubscribeTx allows to subscribe on transaction events by id
	SubscribeTx(ctx context.Context, channelName string, tx ChaincodeTx) TxSubscription
	// SubscribeBlock allows to subscribe on block events
	SubscribeBlock(ctx context.Context, channelName string, seekOpt ...EventCCSeekOption) BlockSubscription
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

// EventCCSubscription describes chaincode events subscription
type EventCCSubscription interface {
	// Events initiates internal GRPC stream and returns channel on chaincode events or error if stream is failed
	Events() (chan *peer.ChaincodeEvent, error)
	// Errors returns errors associated with this subscription
	Errors() chan error
	// Close terminates internal GRPC stream
	Close() error
}

// EventCCSubscription describes tx subscription
type TxSubscription interface {
	Result() (chan TxEvent, error)
	Close() error
}

type BlockSubscription interface {
	Blocks() (chan *common.Block, error)
	Errors() chan error
	Close() error
}

type TxEvent struct {
	TxId    ChaincodeTx
	Success bool
	Error   error
}

// GRPCStreamError contains original error from GRPC stream
type GRPCStreamError struct {
	Err error
}

func (e *GRPCStreamError) Error() string {
	return fmt.Sprintf("grpc stream error: %s", e.Err)
}

type EnvelopeParsingError struct {
	Err error
}

func (e *EnvelopeParsingError) Error() string {
	return fmt.Sprintf("envelope parsing error: %s", e.Err)
}

type UnknownEventTypeError struct {
	Type string
}

func (e *UnknownEventTypeError) Error() string {
	return fmt.Sprintf("unknown event type: %s", e.Type)
}

type InvalidTxError struct {
	TxId ChaincodeTx
	Code peer.TxValidationCode
}

func (e *InvalidTxError) Error() string {
	return fmt.Sprintf("invalid tx: %s with validation code: %s", e.TxId, e.Code.String())
}
