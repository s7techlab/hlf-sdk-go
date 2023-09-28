package observer

import (
	"context"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	hlfproto "github.com/s7techlab/hlf-sdk-go/proto"
)

type (
	Block struct {
		Block         *hlfproto.Block // parsed block
		BlockOriginal *hlfproto.Block // here is original block before transformation if it is, otherwise it's nil
		Channel       string
		Error         error
		CommonBlock *common.Block
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
