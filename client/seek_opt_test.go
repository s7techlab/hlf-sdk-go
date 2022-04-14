package client_test

import (
	"context"
	"math"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client"
)

func TestSeek(t *testing.T) {

	type test struct {
		blockRange []int64
		want       api.EventCCSeekOption
	}

	var channelHeight uint64 = 5

	tests := []test{
		//default - start from new block
		{blockRange: nil, want: api.SeekNewest()},

		// from the oldest block to maxBlock
		{blockRange: []int64{0}, want: api.SeekOldest()},
		// from specified block
		{blockRange: []int64{4}, want: api.SeekRange(4, math.MaxUint64)},
		// -3 blocks back from channel height
		{blockRange: []int64{-3}, want: api.SeekRange(channelHeight-3, math.MaxUint64)},

		// from the oldest block to current channel height
		{blockRange: []int64{0, 0}, want: api.SeekRange(0, channelHeight)},
		// -3 blocks back from channel height to current channel height
		{blockRange: []int64{-3, 0}, want: api.SeekRange(channelHeight-3, channelHeight)},
		// -2 blocks back from channel height to -1 block from channel height
		{blockRange: []int64{-2, -1}, want: api.SeekRange(channelHeight-2, channelHeight-1)},
		// -2 blocks back from channel height to block 4
		{blockRange: []int64{-2, 4}, want: api.SeekRange(channelHeight-2, 4)},
		// from first block if offset greater than channel height
		{blockRange: []int64{-10, 4}, want: api.SeekRange(0, 4)},
		// from first block to -2 block from channel height
		{blockRange: []int64{-10, -2}, want: api.SeekRange(0, 3)},
		// from first block to block 7
		{blockRange: []int64{-10, 7}, want: api.SeekRange(0, 7)},
	}

	logger, _ := zap.NewDevelopment()
	opt := &client.SeekOptConverter{
		GetChannelHeight: func(ctx context.Context, channel string) (uint64, error) {
			return channelHeight, nil
		},
		Logger: logger,
	}

	ctx := context.Background()
	channel := `channel.name`

	for pos, tc := range tests {
		got, err := opt.ByBlockRange(ctx, channel, tc.blockRange...)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		gotSeekFrom, gotSeekTo := got()
		expectedSeekFrom, expectedSeekTo := tc.want()

		if !reflect.DeepEqual(expectedSeekFrom, gotSeekFrom) {
			t.Fatalf("%d. seek from: expected= %v, got= %v", pos, expectedSeekFrom, gotSeekFrom)
		}

		if !reflect.DeepEqual(expectedSeekTo, gotSeekTo) {
			t.Fatalf("%d. seek to: expected= %v, got= %v", pos, expectedSeekTo, gotSeekTo)
		}
	}
}
