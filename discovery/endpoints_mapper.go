package discovery

import (
	"sync"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/api/config"
)

// implementation of connectionMapper interface
var _ connectionMapper = (*EndpointsMapper)(nil)

// EndpointsMapper - if tls is enabled with gossip maps provided from cfg TLS certs to discovered peers
type EndpointsMapper struct {
	addressEndpoint map[string]*api.Endpoint
	lock            sync.RWMutex
}

func NewEndpointsMapper(endpoints []config.Endpoint) *EndpointsMapper {
	endpointMap := make(map[string]*api.Endpoint)

	for _, e := range endpoints {
		var hostAddress api.Endpoint
		hostAddress.TlsConfig = e.TlsConfig

		hostAddress.Host = e.Host
		if e.HostOverride != "" {
			hostAddress.Host = e.HostOverride
		}

		endpointMap[e.Host] = &hostAddress
	}

	return &EndpointsMapper{
		addressEndpoint: endpointMap,
		lock:            sync.RWMutex{},
	}
}

func (m *EndpointsMapper) MapConnection(address string) *api.Endpoint {
	m.lock.RLock()
	defer m.lock.RUnlock()

	endpoint, ok := m.addressEndpoint[address]
	if ok {
		return endpoint
	}

	return &api.Endpoint{}
}

// TlsConfigForAddress - get tls config for provided address
// if config wasn't provided on startup time return disabled tls
func (m *EndpointsMapper) TlsConfigForAddress(address string) config.TlsConfig {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, ok := m.addressEndpoint[address]
	if ok {
		return v.TlsConfig
	}

	return config.TlsConfig{
		Enabled: false,
	}
}

func (m *EndpointsMapper) TlsEndpointForAddress(address string) string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, ok := m.addressEndpoint[address]
	if ok {
		return v.Host
	}

	return address
}

/*
	decorators over api.ChaincodeDiscoverer/ChannelDiscoverer
	adds TLS settings(if they were provided in cfg) for discovered peers
*/

type chaincodeDiscovererTLSDecorator struct {
	target    api.ChaincodeDiscoverer
	tlsMapper connectionMapper
}

func newChaincodeDiscovererTLSDecorator(
	target api.ChaincodeDiscoverer,
	tlsMapper connectionMapper,
) *chaincodeDiscovererTLSDecorator {
	return &chaincodeDiscovererTLSDecorator{
		target:    target,
		tlsMapper: tlsMapper,
	}
}

func (d *chaincodeDiscovererTLSDecorator) Endorsers() []*api.HostEndpoint {
	return addTLConfigs(d.target.Endorsers(), d.tlsMapper)
}

func (d *chaincodeDiscovererTLSDecorator) Orderers() []*api.HostEndpoint {
	return addTLConfigs(d.target.Orderers(), d.tlsMapper)
}

func (d *chaincodeDiscovererTLSDecorator) ChaincodeVersion() string {
	return d.target.ChaincodeVersion()
}

func (d *chaincodeDiscovererTLSDecorator) ChaincodeName() string {
	return d.target.ChaincodeName()
}

func (d *chaincodeDiscovererTLSDecorator) ChannelName() string {
	return d.target.ChannelName()
}

/* */
type channelDiscovererTLSDecorator struct {
	target    api.ChannelDiscoverer
	tlsMapper connectionMapper
}

func newChannelDiscovererTLSDecorator(
	target api.ChannelDiscoverer,
	tlsMapper connectionMapper,
) *channelDiscovererTLSDecorator {
	return &channelDiscovererTLSDecorator{
		target:    target,
		tlsMapper: tlsMapper,
	}
}

func (d *channelDiscovererTLSDecorator) Orderers() []*api.HostEndpoint {
	return addTLConfigs(d.target.Orderers(), d.tlsMapper)
}

func (d *channelDiscovererTLSDecorator) ChannelName() string {
	return d.target.ChannelName()
}

func addTLConfigs(endpoints []*api.HostEndpoint, tlsMapper connectionMapper) []*api.HostEndpoint {
	for i := range endpoints {
		for j := range endpoints[i].HostAddresses {
			conn := tlsMapper.MapConnection(endpoints[i].HostAddresses[j].Host)

			endpoints[i].HostAddresses[j].TlsConfig = conn.TlsConfig
			endpoints[i].HostAddresses[j].Host = conn.Host
		}
	}
	return endpoints
}

/* */
type localPeersDiscovererTLSDecorator struct {
	target    api.LocalPeersDiscoverer
	tlsMapper connectionMapper
}

func newLocalPeersDiscovererTLSDecorator(
	target api.LocalPeersDiscoverer,
	tlsMapper connectionMapper,
) *localPeersDiscovererTLSDecorator {
	return &localPeersDiscovererTLSDecorator{
		target:    target,
		tlsMapper: tlsMapper,
	}
}

func (d *localPeersDiscovererTLSDecorator) Peers() []*api.HostEndpoint {
	return addTLConfigs(d.target.Peers(), d.tlsMapper)
}
