package api

import (
	"context"
	"fmt"
	"math"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc/codes"
)

var (
	SeekFromOldest = &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Oldest{Oldest: &orderer.SeekOldest{}}}
	SeekFromNewest = &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{Newest: &orderer.SeekNewest{}}}
	SeekToMax = NewSeekSpecified(math.MaxUint64)
)

type DeliverClient interface {
	// SubscribeCC allows subscribing on chaincode events using name of channel, chaincode and block offset
	SubscribeCC(ctx context.Context, channelName string, ccName string, seekOpt ...EventCCSeekOption) (EventCCSubscription, error)
	// SubscribeTx allows subscribing on transaction events by id
	SubscribeTx(ctx context.Context, channelName string, txID string, seekOpt ...EventCCSeekOption) (TxSubscription, error)
	// SubscribeBlock allows subscribing on block events. Always returns new instance of block subscription
	SubscribeBlock(ctx context.Context, channelName string, seekOpt ...EventCCSeekOption) (BlockSubscription, error)
}

type EventCCSeekOption func() (*orderer.SeekPosition, *orderer.SeekPosition)

// SeekNewest sets offset to new channel blocks
func SeekNewest() EventCCSeekOption {
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		return SeekFromNewest, SeekToMax
	}
}

// SeekOldest sets offset to channel blocks from beginning
func SeekOldest() EventCCSeekOption {
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		return SeekFromOldest, SeekToMax
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
	var seekFrom *orderer.SeekPosition
	if start == 0 {
		seekFrom = SeekFromOldest
	} else {
		seekFrom = NewSeekSpecified(start)
	}
	return func() (*orderer.SeekPosition, *orderer.SeekPosition) {
		return seekFrom, NewSeekSpecified(end)
	}
}

type EventCCSubscription interface {
	// Events initiates internal GRPC stream and returns channel on chaincode events
	Events() chan *peer.ChaincodeEvent

	EventsExtended() chan interface {
		Event() *peer.ChaincodeEvent
		Block() uint64
		TxTimestamp() *timestamp.Timestamp
	}
	// Errors returns errors associated with this subscription
	Errors() chan error
	// Close cancels current subscription
	Close() error
}

// TxSubscription describes tx subscription
type TxSubscription interface {
	// Result returns result of current tx: success flag, original peer validation code and error if occurred
	Result() (peer.TxValidationCode, error)
	Close() error
}

type BlockSubscription interface {
	Blocks() <-chan *common.Block
	// DEPRECATED: will migrate to just once Err() <- chan error
	Errors() chan error
	Close() error
}

type TxEvent struct {
	TxId    string
	Success bool
	Error   error
}

// GRPCStreamError contains original error from GRPC stream
type GRPCStreamError struct {
	Code codes.Code
	Err  error
}

func (e GRPCStreamError) Error() string {
	return fmt.Sprintf("grpc stream error: %s", e.Err)
}

type EnvelopeParsingError struct {
	Err error
}

func (e EnvelopeParsingError) Error() string {
	return fmt.Sprintf("envelope parsing error: %s", e.Err)
}

type UnknownEventTypeError struct {
	Type string
}

func (e UnknownEventTypeError) Error() string {
	return fmt.Sprintf("unknown event type: %s", e.Type)
}

type InvalidTxError struct {
	TxId string
	Code peer.TxValidationCode
}

func (e InvalidTxError) Error() string {
	return fmt.Sprintf("invalid tx: %s with validation code: %s", e.TxId, e.Code.String())
}
