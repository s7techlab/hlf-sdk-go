package observer

import (
	"context"
	"time"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type (
	ParsedBlock struct {
		Block         *hlfproto.Block // parsed block
		BlockOriginal *hlfproto.Block // here is original block before transformation if it is, otherwise it's nil
		Channel       string
		Error         error
	}

	CreateParsedBlockStream func(context.Context) (<-chan *hlfproto.Block, error)

	CreateParsedBlockStreamWithRetry func(context.Context, CreateParsedBlockStream) (<-chan *hlfproto.Block, error)
)

func CreateParsedBlockStreamWithRetryDelay(delay time.Duration) CreateParsedBlockStreamWithRetry {
	return func(ctx context.Context, createParsedBlockStream CreateParsedBlockStream) (<-chan *hlfproto.Block, error) {
		for {
			select {
			case <-ctx.Done():
				return nil, nil
			default:
			}

			blocks, err := createParsedBlockStream(ctx)
			if err == nil {
				return blocks, nil
			}

			time.Sleep(delay)
		}
	}
}
