package discovery

import "github.com/s7techlab/hlf-sdk-go/api"

// implementation of api.ChaincodeDiscoverer interface
var _ api.ChaincodeDiscoverer = (*chaincodeDTO)(nil)

// chaincodeDTO - chaincode data storage
type chaincodeDTO struct {
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
		chaincodeName:    ccName,
		chaincodeVersion: ccVer,
		channelName:      chanName,
	}
}

func (d *chaincodeDTO) Endorsers() []*api.HostEndpoint {
	return mapToArray(d.endorsers)
}
func (d *chaincodeDTO) Orderers() []*api.HostEndpoint {
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
	d.endorsers[mspID] = append(d.endorsers[mspID], hostAddr)
}

func (d *chaincodeDTO) addEndpointToOrderers(mspID, hostAddr string) {
	d.orderers[mspID] = append(d.orderers[mspID], hostAddr)
}

func (d *chaincodeDTO) addEndpointToPeers(mspID, hostAddr string) {
	d.peers[mspID] = append(d.peers[mspID], hostAddr)
}

func mapToArray(hosts map[string][]string) []*api.HostEndpoint {
	res := make([]*api.HostEndpoint, 0)
	for k := range hosts {
		endpoints := hosts[k]

		he := &api.HostEndpoint{
			MspID:         k,
			HostAddresses: endpoints,
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
	orderers    map[string][]string
	channelName string
}

func newChannelDTO(chanName string) *channelDTO {
	return &channelDTO{
		channelName: chanName,
	}
}

func (d *channelDTO) Orderers() []*api.HostEndpoint {
	return mapToArray(d.orderers)
}

func (d *channelDTO) ChannelName() string {
	return d.channelName
}

func (d *channelDTO) addEndpointToOrderers(mspID, hostAddr string) {
	d.orderers[mspID] = append(d.orderers[mspID], hostAddr)
}
