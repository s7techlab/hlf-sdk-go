package discovery

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
)

var (
	providerStore = make(map[string]api.DiscoveryProvider)
	providerMx    sync.Mutex

	ErrNoChannels      = errors.New(`channels not found`)
	ErrChannelNotFound = errors.New(`channel not found`)
	ErrNoChaincodes    = errors.New(`no chaincodes on channel`)
	ErrNoEndorsers     = errors.New(`endorsers not found`)
	ErrUnknownProvider = errors.New(`unknown discovery provider (forgotten import?)`)
)

func SetProvider(name string, p api.DiscoveryProvider) {
	providerMx.Lock()
	defer providerMx.Unlock()
	providerStore[name] = p
}

func GetProvider(name string) (api.DiscoveryProvider, error) {
	providerMx.Lock()
	defer providerMx.Unlock()
	if p, ok := providerStore[name]; ok {
		return p, nil
	}
	return nil, ErrUnknownProvider
}
