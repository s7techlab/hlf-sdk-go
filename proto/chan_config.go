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
	// todo can be taken from lib
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
)

type ChannelConfig struct {
	Applications map[string]ApplicationConfig `json:"applications"`
	Orderers     map[string]OrdererConfig     `json:"orderers"`

	OrdererBatchSize    orderer.BatchSize     `json:"orderer_batch_size"`
	OrdererBatchTimeout string                `json:"orderer_batch_timeout"`
	OrdererConsesusType orderer.ConsensusType `json:"orderer_consensus_type"`

	Consortium                string                           `json:"consortium"`
	HashingAlgorithm          string                           `json:"hashing_algorithm"`
	BlockDataHashingStructure common.BlockDataHashingStructure `json:"block_data_hashing_structure"`
	Capabilities              common.Capabilities              `json:"capabilities"`
}

type ApplicationConfig struct {
	Name        string             `json:"name"`
	MSP         MSP                `json:"msp"`
	AnchorPeers []*peer.AnchorPeer `json:"anchor_peers"`
}

type MSP struct {
	Config msp.FabricMSPConfig
	// todo Policies
}

type OrdererConfig struct {
	Name      string   `json:"name"`
	MSP       MSP      `json:"msp"`
	Endpoints []string `json:"endpoints"`
}

func ParseChannelConfig(cc common.Config) (*ChannelConfig, error) {
	chanCfg := &ChannelConfig{}

	appCfg, err := ParseApplicationConfig(cc)
	if err != nil {
		return nil, fmt.Errorf("parse application config: %w", err)
	}
	chanCfg.Applications = appCfg

	orderers, err := ParseOrderer(cc)
	if err != nil {
		return nil, fmt.Errorf("parse orderers config: %w", err)
	}
	chanCfg.Orderers = orderers

	batchSize, err := ParseOrdererBatchSize(cc)
	if err != nil {
		return nil, fmt.Errorf("parse batch size: %w", err)
	}
	chanCfg.OrdererBatchSize = *batchSize

	batchTimeout, err := ParseOrdererBatchTimeout(cc)
	if err != nil {
		return nil, fmt.Errorf("parse batch timeout: %w", err)
	}
	chanCfg.OrdererBatchTimeout = batchTimeout

	consensusType, err := ParseOrdererConsesusType(cc)
	if err != nil {
		return nil, fmt.Errorf("parse consensus type: %w", err)
	}
	chanCfg.OrdererConsesusType = *consensusType

	consortium, err := ParseConsortium(cc)
	if err != nil {
		return nil, fmt.Errorf("parse consortium: %w", err)
	}
	chanCfg.Consortium = consortium

	hashingAlgorithm, err := ParseHashingAlgorithm(cc)
	if err != nil {
		return nil, fmt.Errorf("parse hashing algorithm: %w", err)
	}
	chanCfg.HashingAlgorithm = hashingAlgorithm

	blockDataHashing, err := ParseBlockDataHashingStructure(cc)
	if err != nil {
		return nil, fmt.Errorf("parse block data hashing structure: %w", err)
	}
	chanCfg.BlockDataHashingStructure = *blockDataHashing

	capabilities, err := ParseCapabilities(cc)
	if err != nil {
		return nil, fmt.Errorf("parse capabilities: %w", err)
	}
	chanCfg.Capabilities = *capabilities

	return chanCfg, nil
}

func ParseApplicationConfig(cfg common.Config) (map[string]ApplicationConfig, error) {
	applicationGroup, exists := cfg.ChannelGroup.Groups[applicationKey]
	if !exists {
		return nil, fmt.Errorf("application group doesn't exists")
	}

	appCfg := map[string]ApplicationConfig{}

	for groupName := range applicationGroup.Groups {
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

func ParseMSP(b []byte) (*MSP, error) {
	mspCfg := &msp.MSPConfig{}
	if err := proto.Unmarshal(b, mspCfg); err != nil {
		return nil, fmt.Errorf("unmarshal MSPConfig: %w", err)
	}

	fmspCfg := &msp.FabricMSPConfig{}
	if err := proto.Unmarshal(mspCfg.Config, fmspCfg); err != nil {
		return nil, fmt.Errorf("unmarshal FabricMSPConfig: %w", err)
	}

	return &MSP{Config: *fmspCfg}, nil
}

func ParseAnchorPeers(b []byte) ([]*peer.AnchorPeer, error) {
	anchorPeers := &peer.AnchorPeers{}
	if err := proto.Unmarshal(b, anchorPeers); err != nil {
		return nil, fmt.Errorf("unmarshal anchor peers: %w", err)
	}

	return anchorPeers.AnchorPeers, nil
}

func ParseOrderer(cfg common.Config) (map[string]OrdererConfig, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[ordererKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", ordererKey)
	}
	orderersCfg := map[string]OrdererConfig{}

	for groupName := range ordererGroup.Groups {
		mspCV, ok := ordererGroup.Groups[groupName].Values[mspKey]
		if !ok {
			return nil, fmt.Errorf("%v type group doesn't exists", mspKey)
		}

		mspCfg, err := ParseMSP(mspCV.Value)
		if err != nil {
			return nil, fmt.Errorf("parse msp: %w", err)
		}

		endpointsCV, ok := ordererGroup.Groups[groupName].Values[endpointsKey]
		if !ok {
			return nil, fmt.Errorf("%v type group doesn't exists", endpointsKey)
		}

		endpoints, err := ParseOrdererEndpoints(endpointsCV.Value)
		if err != nil {
			return nil, fmt.Errorf("parse endpoints: %w", err)
		}

		orderersCfg[groupName] = OrdererConfig{
			Name:      groupName,
			MSP:       *mspCfg,
			Endpoints: endpoints,
		}
	}

	return orderersCfg, nil
}

func ParseOrdererEndpoints(b []byte) ([]string, error) {
	oa := &common.OrdererAddresses{}
	if err := proto.Unmarshal(b, oa); err != nil {
		return nil, fmt.Errorf("unmarshal OrdererAddresses: %w", err)
	}

	return oa.Addresses, nil
}

//
func ParseOrdererBatchSize(cfg common.Config) (*orderer.BatchSize, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[ordererKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", ordererKey)
	}

	batchSize, exists := ordererGroup.Values[batchSizeKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", batchSizeKey)
	}

	return ParseBatchSizeFromBytes(batchSize.Value)
}

func ParseBatchSizeFromBytes(b []byte) (*orderer.BatchSize, error) {
	bs := &orderer.BatchSize{}
	if err := proto.Unmarshal(b, bs); err != nil {
		return nil, fmt.Errorf("unmarshal BatchSize: %w", err)
	}

	return bs, nil
}

//
func ParseOrdererBatchTimeout(cfg common.Config) (string, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[ordererKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", ordererKey)
	}

	batchTimeout, exists := ordererGroup.Values[batchTimeoutKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", batchTimeoutKey)
	}

	return ParseOrdererBatchTimeoutFromBytes(batchTimeout.Value)
}

func ParseOrdererBatchTimeoutFromBytes(b []byte) (string, error) {
	bt := &orderer.BatchTimeout{}
	if err := proto.Unmarshal(b, bt); err != nil {
		return "", fmt.Errorf("unmarshal BatchTimeout: %w", err)
	}
	return bt.Timeout, nil
}

//
func ParseOrdererConsesusType(cfg common.Config) (*orderer.ConsensusType, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[ordererKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", ordererKey)
	}

	consensusType, exists := ordererGroup.Values[consensusTypeKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", consensusTypeKey)
	}

	return ParseOrdererConsesusTypeFromBytes(consensusType.Value)
}

func ParseOrdererConsesusTypeFromBytes(b []byte) (*orderer.ConsensusType, error) {
	ct := &orderer.ConsensusType{}
	if err := proto.Unmarshal(b, ct); err != nil {
		return nil, fmt.Errorf("unmarshal ConsensusType: %w", err)
	}
	return ct, nil
}

//
func ParseConsortium(cfg common.Config) (string, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[ordererKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", ordererKey)
	}

	consensusType, exists := ordererGroup.Values[consensusTypeKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", consensusTypeKey)
	}

	return ParseConsortiumFromBytes(consensusType.Value)
}

func ParseConsortiumFromBytes(b []byte) (string, error) {
	c := &common.Consortium{}
	if err := proto.Unmarshal(b, c); err != nil {
		return "", fmt.Errorf("unmarshal Consortium: %w", err)
	}
	return c.Name, nil
}

//
func ParseHashingAlgorithm(cfg common.Config) (string, error) {
	hashingAlgorithm, exists := cfg.ChannelGroup.Values[hashingAlgorithmKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", hashingAlgorithmKey)
	}

	return ParseHashingAlgorithmFromBytes(hashingAlgorithm.Value)
}

func ParseHashingAlgorithmFromBytes(b []byte) (string, error) {
	ha := &common.HashingAlgorithm{}
	if err := proto.Unmarshal(b, ha); err != nil {
		return "", fmt.Errorf("unmarshal HashingAlgorithm: %w", err)
	}
	return ha.Name, nil
}

//
func ParseBlockDataHashingStructure(cfg common.Config) (*common.BlockDataHashingStructure, error) {
	bdh, exists := cfg.ChannelGroup.Values[blockDataHashingStructureKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", blockDataHashingStructureKey)
	}

	return ParseParseBlockDataHashingStructureFromBytes(bdh.Value)
}

func ParseParseBlockDataHashingStructureFromBytes(b []byte) (*common.BlockDataHashingStructure, error) {
	bdh := &common.BlockDataHashingStructure{}
	if err := proto.Unmarshal(b, bdh); err != nil {
		return nil, fmt.Errorf("unmarshal BatchTimeout: %w", err)
	}
	return bdh, nil
}

//
func ParseCapabilities(cfg common.Config) (*common.Capabilities, error) {
	bdh, exists := cfg.ChannelGroup.Values[channelconfig.CapabilitiesKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.CapabilitiesKey)
	}

	return ParseParseCapabilitiesFromBytes(bdh.Value)
}

func ParseParseCapabilitiesFromBytes(b []byte) (*common.Capabilities, error) {
	c := &common.Capabilities{}
	if err := proto.Unmarshal(b, c); err != nil {
		return nil, fmt.Errorf("unmarshal BatchTimeout: %w", err)
	}
	return c, nil
}
