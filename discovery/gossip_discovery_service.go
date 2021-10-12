package discovery

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/discovery"
	discClient "github.com/hyperledger/fabric/discovery/client"
)

// gossipServiceDiscovery - fetches info about all available peers, endorsers and orderers for channel & chaincode
// via configured gossip protocol
// helper module for GossipDiscoveryProvider
type gossipServiceDiscovery struct {
	client         *discClient.Client
	clientIdentity []byte
}

func newGossipServiceDiscovery(client *discClient.Client, clientIdentity []byte) *gossipServiceDiscovery {
	return &gossipServiceDiscovery{
		client:         client,
		clientIdentity: clientIdentity,
	}
}

// DiscoverChaincode - find available peers, endorsers and orderers for channel & chaincode
func (s *gossipServiceDiscovery) DiscoverChaincode(ctx context.Context, ccName, chanName string) (*chaincodeDTO, error) {
	// TODO err handling/diversification
	req, err := discClient.
		NewRequest().
		OfChannel(chanName).
		AddPeersQuery().
		AddConfigQuery().
		AddEndorsersQuery(&discovery.ChaincodeInterest{
			Chaincodes: []*discovery.ChaincodeCall{
				{Name: ccName},
			},
		})
	if err != nil {
		return nil, err
	}

	res, err := s.client.Send(ctx, req, s.getAuthInfo())
	if err != nil {
		return nil, err
	}

	chanEndorsers, err := res.ForChannel(chanName).Endorsers(discClient.InvocationChain{
		&discovery.ChaincodeCall{
			Name: ccName,
		},
	}, discClient.NoFilter)
	if err != nil {
		return nil, err
	}

	chanPeers, err := res.ForChannel(chanName).Peers()
	if err != nil {
		return nil, err
	}
	chanCfg, err := res.ForChannel(chanName).Config()
	if err != nil {
		return nil, err
	}

	// get chaincode version
	var ccVersion string
	if len(chanEndorsers) != 0 {
		ccodes := chanEndorsers[0].StateInfoMessage.GossipMessage.GetStateInfo().Properties.Chaincodes
		if len(ccodes) != 0 {
			ccVersion = ccodes[0].Version
		}
	}

	dc := newChaincodeDTO(ccName, ccVersion, chanName)
	return s.parseDiscoverChaincodeResponse(dc, chanEndorsers, chanPeers, chanCfg), nil
}

// DiscoverChannel - returns orderers for provided channel
func (s *gossipServiceDiscovery) DiscoverChannel(ctx context.Context, chanName string) (*channelDTO, error) {
	req := discClient.
		NewRequest().
		OfChannel(chanName).
		AddConfigQuery()

	res, err := s.client.Send(ctx, req, s.getAuthInfo())
	if err != nil {
		return nil, err
	}

	chanCfg, err := res.ForChannel(chanName).Config()
	if err != nil {
		return nil, err
	}

	dc := newChannelDTO(chanName)

	return s.parseDiscoverChannelResponse(dc, chanCfg), nil
}

func (s *gossipServiceDiscovery) parseDiscoverChaincodeResponse(
	dc *chaincodeDTO,
	endorsers discClient.Endorsers,
	peers []*discClient.Peer,
	cfg *discovery.ConfigResult,
) *chaincodeDTO {
	for i := range endorsers {
		hostAddr := endorsers[i].AliveMessage.GetAliveMsg().Membership.Endpoint
		dc.addEndpointToEndorsers(endorsers[i].MSPID, hostAddr)
	}

	for i := range peers {
		hostAddr := peers[i].AliveMessage.GetAliveMsg().Membership.Endpoint
		dc.addEndpointToEndorsers(peers[i].MSPID, hostAddr)
	}

	for ordererMSPID := range cfg.Orderers {
		for i := range cfg.Orderers[ordererMSPID].Endpoint {
			hostAddr := fmt.Sprintf("%s:%d", cfg.Orderers[ordererMSPID].Endpoint[i].Host, cfg.Orderers[ordererMSPID].Endpoint[i].Port)
			dc.addEndpointToOrderers(ordererMSPID, hostAddr)
		}
	}

	return dc
}

func (s *gossipServiceDiscovery) parseDiscoverChannelResponse(
	dc *channelDTO,
	cfg *discovery.ConfigResult,
) *channelDTO {
	for ordererMSPID := range cfg.Orderers {
		for i := range cfg.Orderers[ordererMSPID].Endpoint {
			hostAddr := fmt.Sprintf("%s:%d", cfg.Orderers[ordererMSPID].Endpoint[i].Host, cfg.Orderers[ordererMSPID].Endpoint[i].Port)
			dc.addEndpointToOrderers(ordererMSPID, hostAddr)
		}
	}

	return dc
}

func (s *gossipServiceDiscovery) getAuthInfo() *discovery.AuthInfo {
	return &discovery.AuthInfo{
		ClientIdentity: s.clientIdentity,
	}
}
