package observer

import (
	"context"
	"time"
)

type (
	CreateBlockStream[T any] func(context.Context) (<-chan T, error)

	CreateBlockStreamWithRetry[T any] func(context.Context, CreateBlockStream[T]) (<-chan T, error)
)

func CreateBlockStreamWithRetryDelay[T any](delay time.Duration) CreateBlockStreamWithRetry[T] {
	return func(ctx context.Context, createBlockStream CreateBlockStream[T]) (<-chan T, error) {
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
