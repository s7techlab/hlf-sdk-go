package service_discovery

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/discovery"
	discClient "github.com/hyperledger/fabric/discovery/client"
)

type ServiceDiscovery struct {
	client         *discClient.Client
	clientIdentity []byte
}

func NewServiceDiscovery(client *discClient.Client, clientIdentity []byte) *ServiceDiscovery {
	return &ServiceDiscovery{
		client:         client,
		clientIdentity: clientIdentity,
	}
}

// GetEndorsers - endorser peers from MSPs
func (s *ServiceDiscovery) GetEndorsers(ctx context.Context, ccName, chanName string) (discClient.Endorsers, error) {
	req, err := discClient.
		NewRequest().
		OfChannel(chanName).
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
			//CollectionNames: cc.Collections,
		},
	}, discClient.NoFilter)
	if err != nil {
		return nil, err
	}

	return chanEndorsers, nil
}

// GetPeers - returns all peers connected to channel
func (s *ServiceDiscovery) GetPeers(ctx context.Context, chanName string) ([]*discClient.Peer, error) {
	req := discClient.
		NewRequest().
		OfChannel(chanName).
		AddPeersQuery()

	res, err := s.client.Send(ctx, req, s.getAuthInfo())
	if err != nil {
		return nil, err
	}

	chanPeers, err := res.ForChannel(chanName).Peers()
	if err != nil {
		return nil, err
	}

	return chanPeers, nil
}

// GetChannelConfig - return certs(root, tls,) of MSPs and Orderers
func (s *ServiceDiscovery) GetChannelConfig(ctx context.Context, chanName string) (*discovery.ConfigResult, error) {
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

	return chanCfg, nil
}

func (s *ServiceDiscovery) getAuthInfo() *discovery.AuthInfo {
	return &discovery.AuthInfo{
		ClientIdentity: s.clientIdentity,
		// TODO is it necessary? ClientTlsCertHash: ,
	}
}
