package discovery

import (
	"github.com/pkg/errors"
)

var (
	ErrNoChannels      = errors.New(`channels not found`)
	ErrChannelNotFound = errors.New(`channel not found`)
	ErrNoChaincodes    = errors.New(`no chaincodes on channel`)
	ErrUnknownProvider = errors.New(`unknown discovery provider (forgotten import?)`)
)

// ServiceDiscoveryType - what types of discovery we support
type ServiceDiscoveryType string

const (
	LocalConfigServiceDiscoveryType ServiceDiscoveryType = "local"
	GossipServiceDiscoveryType      ServiceDiscoveryType = "gossip"
)
