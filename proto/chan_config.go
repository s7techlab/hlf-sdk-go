package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/channelconfig"
)

const (
	applicationKey               = "Application"
	ordererKey                   = "Orderer"
	mspKey                       = "MSP"
	endpointsKey                 = "Endpoints"
	batchSizeKey                 = "BatchSize"
	batchTimeoutKey              = "BatchTimeout"
	consensusTypeKey             = "ConsensusType"
	consortiumKey                = "Consortium"
	hashingAlgorithmKey          = "HashingAlgorithm"
	ordererAddressesKey          = "OrdererAddresses"
	blockDataHashingStructureKey = "BlockDataHashingStructure"

	caCertType           = "ca"
	intermediateCertType = "intermediate"
	adminCertType        = "admin"
)

type ChannelConfig struct {
	Applications map[string]ApplicationConfig `json:"applications"`
	Orderers     map[string]OrdererConfig     `json:"orderers"`

	OrdererBatchSize    orderer.BatchSize     `json:"orderer_batch_size"`
	OrdererBatchTimeout string                `json:"orderer_batch_timeout"`
	OrdererConsesusType orderer.ConsensusType `json:"orderer_consensus_type"`

	Consortium                string                           `json:"consortium"`
	HashingAlgorithm          string                           `json:"hashing_algorithm"`
	OrdererAddresses          []string                         `json:"orderer_addresses"`
	BlockDataHashingStructure common.BlockDataHashingStructure `json:"block_data_hashing_structure"`
}

type ApplicationConfig struct {
	Name        string             `json:"name"`
	MSP         MSP                `json:"msp"`
	AnchorPeers []*peer.AnchorPeer `json:"anchor_peers"`
}

type MSP struct {
	Config msp.FabricMSPConfig
}

type OrdererConfig struct {
	Name      string   `json:"name"`
	MSP       MSP      `json:"msp"`
	Endpoints []string `json:"endpoints"`
}

type Certificate struct {
	Fingerprint []byte
	Data        []byte
	MSPID       string
	Type        string
	MSPName     string
}

func ParseApplicationConfig(cfg common.Config) (map[string]ApplicationConfig, error) {
	applicationGroup, exists := cfg.ChannelGroup.Groups[applicationKey]
	if !exists {
		return nil, fmt.Errorf("application group doesn't exists")
	}

	appCfg := map[string]ApplicationConfig{}

	for groupName := range applicationGroup.Groups {
		var (
			a ApplicationConfig
		)

		a.Name = groupName
		mspCfg, err := ParseMSP(applicationGroup.Groups[groupName].Values[mspKey].Value)
		if err != nil {
			return nil, fmt.Errorf("parse msp: %w", err)
		}

		ancPeers, err := ParseAnchorPeers(applicationGroup.Groups[groupName].Values[channelconfig.AnchorPeersKey].Value)
		if err != nil {
			return nil, fmt.Errorf("parse anchor peers: %w", err)
		}

		appCfg[groupName] = ApplicationConfig{
			Name:        groupName,
			MSP:         *mspCfg,
			AnchorPeers: ancPeers,
		}
	}
	return appCfg, nil
}

func ParseAnchorPeers(b []byte) ([]*peer.AnchorPeer, error) {
	anchorPeers := &peer.AnchorPeers{}

	if err := proto.Unmarshal(b, anchorPeers); err != nil {
		return nil, fmt.Errorf("unmarshal anchor peers from config: %w", err)
	}

	return anchorPeers.AnchorPeers, nil
}

func ParseMSP(b []byte) (*MSP, error) {
	mspCfg := &msp.MSPConfig{}
	if err := proto.Unmarshal(b, mspCfg); err != nil {
		return nil, fmt.Errorf("JSON unmarshal application MSP config: %w", err)
	}

	fmspCfg := &msp.FabricMSPConfig{}

	if err := proto.Unmarshal(mspCfg.Config, fmspCfg); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal MSP: %w", err)
	}

	return &MSP{Config: *fmspCfg}, nil
}
func ParseChannelConfig(cc common.Config) (*ChannelConfig, error) {
	chanCfg := &ChannelConfig{}

	appCfg, err := ParseApplicationConfig(cc)
	if err != nil {
		return nil, fmt.Errorf("parse application config: %w", err)
	}

	chanCfg.Applications = appCfg

	return chanCfg, nil
}
