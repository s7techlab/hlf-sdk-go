package fetcher

import "context"

// Nope do nothing
type Nope struct {
}

func (n *Nope) Fetch(ctx context.Context, repo, version string) ([]byte, error) {
	return nil, nil
}

func NewNope() *Nope {
	return &Nope{}
}
