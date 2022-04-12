package client

import (
	"context"
	"fmt"

	ordererproto "github.com/hyperledger/fabric-protos-go/orderer"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
)

type SeekOptConverter struct {
	GetChannelHeight func(ctx context.Context, channel string) (uint64, error)
	currentHeight    uint64
	Logger           *zap.Logger
}

func NewSeekOptConverter(c *core) *SeekOptConverter {
	return &SeekOptConverter{
		GetChannelHeight: func(ctx context.Context, channel string) (uint64, error) {
			channelInfo, err := c.System().QSCC().GetChainInfo(ctx, &system.GetChainInfoRequest{ChannelName: channel})
			if err != nil {
				return 0, err
			}

			return channelInfo.Height, nil
		},
		Logger: c.logger,
	}
}

func (so *SeekOptConverter) ChannelHeight(ctx context.Context, channel string) (uint64, error) {
	var err error
	if so.currentHeight != 0 {
		return so.currentHeight, nil
	}

	so.currentHeight, err = so.GetChannelHeight(ctx, channel)
	return so.currentHeight, err
}

func (so *SeekOptConverter) ByBlockRange(ctx context.Context, channel string, blockRange ...int64) (api.EventCCSeekOption, error) {

	var (
		blockRangeFrom, blockRangeTo int64
		seekFrom, seekTo             *ordererproto.SeekPosition
	)

	if blockRange == nil {
		blockRange = []int64{}
	}

	so.Logger.Debug(`seek by block range`, zap.Reflect(`block range`, blockRange))

	switch {
	case len(blockRange) > 0:
		blockRangeFrom = blockRange[0]

		switch {
		// seek from new blocks
		case blockRangeFrom == 0:
			seekFrom = api.SeekFromOldest
		case blockRangeFrom > 0:
			seekFrom = api.SeekSpecified(uint64(blockRangeFrom))
		case blockRangeFrom < 0:
			// from  -{x} means we need to look x blocks back for events
			// thus we need to  know current channel height
			height, err := so.ChannelHeight(ctx, channel)
			if err != nil {
				return nil, fmt.Errorf(`get channel height: %w`, err)
			}
			so.Logger.Debug(`get channel info for calculate negative block from`,
				zap.Uint64(`channel_height`, height))

			seekFrom = api.SeekSpecified(uint64(int64(height) + blockRangeFrom))
		}

	default:
		seekFrom = api.SeekFromNewest
	}

	switch {
	case len(blockRange) > 1:

		blockRangeTo = blockRange[1]

		switch {

		case blockRangeTo > 0:
			seekTo = api.SeekSpecified(uint64(blockRangeTo))

		case blockRangeTo == 0:
			fallthrough

		case blockRangeTo < 0:
			// to  -{x} means we need to look x blocks back to channel height
			// zero means that we receive events until current channel height
			// thus we need to  know current channel height
			height, err := so.ChannelHeight(ctx, channel)
			if err != nil {
				return nil, fmt.Errorf(`get channel height: %w`, err)
			}
			so.Logger.Debug(`get channel info for calculate negative block to`,
				zap.Uint64(`channel_height`, height))

			seekTo = api.SeekSpecified(uint64(int64(height) + blockRangeTo))
		}

	default:
		seekTo = api.SeekToMax
	}

	so.Logger.Debug(`seek opts`,
		zap.Reflect(`seek from`, seekFrom),
		zap.Reflect(`seek to`, seekTo))

	return func() (*ordererproto.SeekPosition, *ordererproto.SeekPosition) {
		return seekFrom, seekTo
	}, nil

}
