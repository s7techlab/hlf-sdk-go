package discovery

import (
	"sync"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
)

// implementation of api.ChaincodeDiscoverer interface
var _ api.ChaincodeDiscoverer = (*chaincodeDTO)(nil)

// chaincodeDTO - chaincode data storage
type chaincodeDTO struct {
	lock sync.RWMutex
	// key - MSPID, value host addresses
	endorsers        map[string][]string
	orderers         map[string][]string
	peers            map[string][]string
	chaincodeName    string
	chaincodeVersion string
	channelName      string
}

func newChaincodeDTO(ccName, ccVer, chanName string) *chaincodeDTO {
	return &chaincodeDTO{
		lock:             sync.RWMutex{},
		chaincodeName:    ccName,
		chaincodeVersion: ccVer,
		channelName:      chanName,
		endorsers:        make(map[string][]string),
		orderers:         make(map[string][]string),
		peers:            make(map[string][]string),
	}
}

func (d *chaincodeDTO) Endorsers() []*api.HostEndpoint {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return mapToArray(d.endorsers)
}
func (d *chaincodeDTO) Orderers() []*api.HostEndpoint {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return mapToArray(d.orderers)
}
func (d *chaincodeDTO) ChaincodeName() string {
	return d.chaincodeName
}
func (d *chaincodeDTO) ChaincodeVersion() string {
	return d.chaincodeVersion
}
func (d *chaincodeDTO) ChannelName() string {
	return d.channelName
}

// helpers
func (d *chaincodeDTO) addEndpointToEndorsers(mspID, hostAddr string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.endorsers[mspID] = append(d.endorsers[mspID], hostAddr)
}

func (d *chaincodeDTO) addEndpointToOrderers(mspID, hostAddr string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.orderers[mspID] = append(d.orderers[mspID], hostAddr)
}

func (d *chaincodeDTO) addEndpointToPeers(mspID, hostAddr string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.peers[mspID] = append(d.peers[mspID], hostAddr)
}

func mapToArray(hosts map[string][]string) []*api.HostEndpoint {
	res := make([]*api.HostEndpoint, 0)
	for k := range hosts {
		endpoints := hosts[k]

		he := &api.HostEndpoint{
			MspID:         k,
			HostAddresses: make([]*api.HostAddress, len(endpoints)),
		}

		for i := range endpoints {
			he.HostAddresses[i] = &api.HostAddress{
				Address: endpoints[i],
			}
		}
		res = append(res, he)
	}

	return res
}

/* */
// implementation of api.ChaincodeDiscoverer interface
var _ api.ChannelDiscoverer = (*channelDTO)(nil)

// channel - info about channel orderers
type channelDTO struct {
	lock        sync.RWMutex
	orderers    map[string][]string
	channelName string
}

func newChannelDTO(chanName string) *channelDTO {
	return &channelDTO{
		lock:        sync.RWMutex{},
		channelName: chanName,
		orderers:    make(map[string][]string),
	}
}

func (d *channelDTO) Orderers() []*api.HostEndpoint {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return mapToArray(d.orderers)
}

func (d *channelDTO) ChannelName() string {
	return d.channelName
}

func (d *channelDTO) addEndpointToOrderers(mspID, hostAddr string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.orderers[mspID] = append(d.orderers[mspID], hostAddr)
}

/* */
// implementation of api.LocalPeersDiscoverer interface
var _ api.LocalPeersDiscoverer = (*localPeersDTO)(nil)

type localPeersDTO struct {
	lock  sync.RWMutex
	peers map[string][]string
}

func newLocalPeersDTO() *localPeersDTO {
	return &localPeersDTO{
		lock:  sync.RWMutex{},
		peers: make(map[string][]string),
	}
}

func (d *localPeersDTO) Peers() []*api.HostEndpoint {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return mapToArray(d.peers)
}

func (d *localPeersDTO) addEndpointToPeers(mspID, hostAddr string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.peers[mspID] = append(d.peers[mspID], hostAddr)
}
