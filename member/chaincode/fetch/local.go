package fetch

import (
	"context"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/s7techlab/hlf-sdk-go/api"
)

type RepoAccess interface {
	Clone(ctx context.Context, s *peer.ChaincodeID) error
}

func NewLocalFetcher(dir string) (api.CCFetcher, error) {
	f := &localFetcher{
		path: dir,
	}

	err := f.check()
	if err != nil {
		return nil, err
	}

	return f, nil
}

type localFetcher struct {
	path string
}

func (f *localFetcher) Fetch(ctx context.Context, id *peer.ChaincodeID) (*peer.ChaincodeDeploymentSpec, error) {
	return nil, nil
}

func (f *localFetcher) check() error {
	return nil
}
