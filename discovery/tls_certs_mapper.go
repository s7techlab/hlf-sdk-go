package discovery

import (
	"sync"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/api/config"
)

// implementation of tlsConfigMapper interface
var _ tlsConfigMapper = (*TLSCertsMapper)(nil)

// TLSCertsMapper - if tls is enabled with gossip maps provided from cfg TLS certs to discovered peers
type TLSCertsMapper struct {
	addrCfgMap map[string]*config.TlsConfig
	lock       sync.RWMutex
}

func NewTLSCertsMapper(certsCfg []config.TLSCertsMapperConfig) *TLSCertsMapper {
	addrCfgMap := map[string]*config.TlsConfig{}

	for i := range certsCfg {
		addrCfgMap[certsCfg[i].Address] = &certsCfg[i].TlsConfig
	}

	return &TLSCertsMapper{addrCfgMap: addrCfgMap, lock: sync.RWMutex{}}
}

// tlsConfigForAddress - get tls config for provided address
// if config wasnt provided on startup time return disabled tls
func (m *TLSCertsMapper) TlsConfigForAddress(address string) *config.TlsConfig {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, ok := m.addrCfgMap[address]
	if ok {
		return v
	}

	return &config.TlsConfig{
		Enabled: false,
	}
}

/*
	decorators over api.ChaincodeDiscoverer/ChannelDiscoverer
	adds TLS settings(if they was provided in cfg) for discovered peers
*/

type chaincodeDiscovererTLSDecorator struct {
	target    api.ChaincodeDiscoverer
	tlsMapper tlsConfigMapper
}

func newChaincodeDiscovererTLSDecorator(
	target api.ChaincodeDiscoverer,
	tlsMapper tlsConfigMapper,
) *chaincodeDiscovererTLSDecorator {
	return &chaincodeDiscovererTLSDecorator{
		target:    target,
		tlsMapper: tlsMapper,
	}
}

func (d *chaincodeDiscovererTLSDecorator) Endorsers() []*api.HostEndpoint {
	return addTLSSettings(d.target.Endorsers(), d.tlsMapper)
}

func (d *chaincodeDiscovererTLSDecorator) Orderers() []*api.HostEndpoint {
	return addTLSSettings(d.target.Orderers(), d.tlsMapper)
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
	tlsMapper tlsConfigMapper
}

func newChannelDiscovererTLSDecorator(
	target api.ChannelDiscoverer,
	tlsMapper tlsConfigMapper,
) *channelDiscovererTLSDecorator {
	return &channelDiscovererTLSDecorator{
		target:    target,
		tlsMapper: tlsMapper,
	}
}

func (d *channelDiscovererTLSDecorator) Orderers() []*api.HostEndpoint {
	return addTLSSettings(d.target.Orderers(), d.tlsMapper)
}

func (d *channelDiscovererTLSDecorator) ChannelName() string {
	return d.target.ChannelName()
}

func addTLSSettings(endpoints []*api.HostEndpoint, tlsMapper tlsConfigMapper) []*api.HostEndpoint {
	for i := range endpoints {
		for j := range endpoints[i].HostAddresses {
			tlsCfg := tlsMapper.TlsConfigForAddress(endpoints[i].HostAddresses[j].Address)
			endpoints[i].HostAddresses[j].TLSSettings = *tlsCfg
		}
	}
	return endpoints
}

/* */

type localPeersDiscovererTLSDecorator struct {
	target    api.LocalPeersDiscoverer
	tlsMapper tlsConfigMapper
}

func newLocalPeersDiscovererTLSDecorator(
	target api.LocalPeersDiscoverer,
	tlsMapper tlsConfigMapper,
) *localPeersDiscovererTLSDecorator {
	return &localPeersDiscovererTLSDecorator{
		target:    target,
		tlsMapper: tlsMapper,
	}
}

func (d *localPeersDiscovererTLSDecorator) Peers() []*api.HostEndpoint {
	return addTLSSettings(d.target.Peers(), d.tlsMapper)
}
