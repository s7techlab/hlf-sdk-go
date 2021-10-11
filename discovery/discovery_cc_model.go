package discovery

import "github.com/s7techlab/hlf-sdk-go/api"

// implementation of api.IDiscoveryChaincode interface
var _ api.IDiscoveryChaincode = (*discoveryChaincode)(nil)

// discoveryChaincode - chaincode data storage
type discoveryChaincode struct {
	// key - MSPID, value host addresses
	endorsers        map[string][]string
	orderers         map[string][]string
	peers            map[string][]string
	chaincodeName    string
	chaincodeVersion string
	channelName      string
}

func newDiscoveryChaincode(ccName, ccVer, chanName string) *discoveryChaincode {
	return &discoveryChaincode{
		chaincodeName:    ccName,
		chaincodeVersion: ccVer,
		channelName:      chanName,
	}
}

func (d *discoveryChaincode) Endorsers() []*api.HostEndpoint {
	return d.mapToArray(d.endorsers)
}
func (d *discoveryChaincode) Orderers() []*api.HostEndpoint {
	return d.mapToArray(d.orderers)
}
func (d *discoveryChaincode) ChaincodeName() string {
	return d.chaincodeName
}
func (d *discoveryChaincode) ChaincodeVersion() string {
	return d.chaincodeVersion
}
func (d *discoveryChaincode) ChannelName() string {
	return d.channelName
}

// helpers
func (d *discoveryChaincode) addEndpointToEndorsers(mspID, hostAddr string) {
	d.endorsers[mspID] = append(d.endorsers[mspID], hostAddr)
}

func (d *discoveryChaincode) addEndpointToOrderers(mspID, hostAddr string) {
	d.orderers[mspID] = append(d.orderers[mspID], hostAddr)
}

func (d *discoveryChaincode) addEndpointToPeers(mspID, hostAddr string) {
	d.peers[mspID] = append(d.peers[mspID], hostAddr)
}

func (d *discoveryChaincode) mapToArray(hosts map[string][]string) []*api.HostEndpoint {
	res := make([]*api.HostEndpoint, 0)
	for k := range d.endorsers {
		endpoints := d.endorsers[k]

		he := &api.HostEndpoint{
			MspID:         k,
			HostAddresses: endpoints,
		}
		res = append(res, he)
	}

	return res
}
