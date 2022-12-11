package fetcher

import "context"

// noneFetcher do nothing
type noneFetcher struct {
}

func (n noneFetcher) Fetch(ctx context.Context, repo, version string) ([]byte, error) {
	return nil, nil
}

func NewNoneFetcher() *noneFetcher {
	return &noneFetcher{}
}
