package observer

import (
	"context"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
)

type (
	Block struct {
		Block   *common.Block
		Channel string
	}

	CreateBlockStream func(context.Context) (<-chan *common.Block, error)

	CreateBlockStreamWithRetry func(context.Context, CreateBlockStream) (<-chan *common.Block, error)
)

func CreateBlockStreamWithRetryDelay(delay time.Duration) CreateBlockStreamWithRetry {
	return func(ctx context.Context, createBlockStream CreateBlockStream) (<-chan *common.Block, error) {
		for {
			select {
			case <-ctx.Done():
				return nil, nil
			default:
			}

			blocks, err := createBlockStream(ctx)
			if err == nil {
				return blocks, nil
			}

			time.Sleep(delay)
		}
	}
}
