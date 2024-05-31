package block

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/channelconfig"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/s7techlab/hlf-sdk-go/proto/block"
)

func UnmarshalChannelConfig(b []byte) (*block.ChannelConfig, error) {
	cc := &block.ChannelConfig{}

	if err := protojson.Unmarshal(b, cc); err != nil {
		return nil, err
	}

	return cc, nil
}

func ParseChannelConfig(cc common.Config) (*block.ChannelConfig, error) {
	chanCfg := &block.ChannelConfig{}

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
	chanCfg.OrdererBatchSize = batchSize

	batchTimeout, err := ParseOrdererBatchTimeout(cc)
	if err != nil {
		return nil, fmt.Errorf("parse batch timeout: %w", err)
	}
	chanCfg.OrdererBatchTimeout = batchTimeout

	consensusType, err := ParseOrdererConsensusType(cc)
	if err != nil {
		return nil, fmt.Errorf("parse consensus type: %w", err)
	}
	chanCfg.OrdererConsensusType = consensusType

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
	chanCfg.BlockDataHashingStructure = blockDataHashing

	capabilities, err := ParseCapabilities(cc)
	if err != nil {
		return nil, fmt.Errorf("parse capabilities: %w", err)
	}
	chanCfg.Capabilities = capabilities

	policies, err := ParsePolicy(cc.ChannelGroup.Policies)
	if err != nil {
		return nil, fmt.Errorf("parse policies: %w", err)
	}

	chanCfg.Policy = policies

	return chanCfg, nil
}

func ParseApplicationConfig(cfg common.Config) (map[string]*block.ApplicationConfig, error) {
	applicationGroup, exists := cfg.ChannelGroup.Groups[channelconfig.ApplicationGroupKey]
	if !exists {
		return nil, fmt.Errorf("application group doesn't exists")
	}

	appCfg := map[string]*block.ApplicationConfig{}

	for groupName := range applicationGroup.Groups {
		var (
			mspCfg   *block.MSP
			ancPeers []*peer.AnchorPeer
			err      error
		)

		mspCfg, err = ParseMSP(applicationGroup.Groups[groupName], groupName)
		if err != nil {
			return nil, fmt.Errorf("parse msp: %w", err)
		}

		ancPeers, err = ParseAnchorPeers(applicationGroup.Groups[groupName])
		if err != nil {
			return nil, fmt.Errorf("parse anchor peers: %w", err)
		}

		appCfg[groupName] = &block.ApplicationConfig{
			Name:        groupName,
			Msp:         mspCfg,
			AnchorPeers: ancPeers,
		}
	}

	return appCfg, nil
}

func ParseMSP(mspConfigGroup *common.ConfigGroup, groupName string) (*block.MSP, error) {
	mspCV, ok := mspConfigGroup.Values[channelconfig.MSPKey]
	if !ok {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.MSPKey)
	}

	mspCfg := &msp.MSPConfig{}
	if err := proto.Unmarshal(mspCV.Value, mspCfg); err != nil {
		return nil, fmt.Errorf("unmarshal MSPConfig: %w", err)
	}

	fabricMSPCfg := &msp.FabricMSPConfig{}
	if err := proto.Unmarshal(mspCfg.Config, fabricMSPCfg); err != nil {
		return nil, fmt.Errorf("unmarshal FabricMSPConfig: %w", err)
	}

	policy, err := ParsePolicy(mspConfigGroup.Policies)
	if err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}

	return &block.MSP{Config: fabricMSPCfg, Policy: policy, Name: groupName}, nil
}

func ParseOrderer(cfg common.Config) (map[string]*block.OrdererConfig, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return nil, nil
	}
	orderersCfg := map[string]*block.OrdererConfig{}

	for groupName := range ordererGroup.Groups {
		mspCfg, err := ParseMSP(ordererGroup.Groups[groupName], groupName)
		if err != nil {
			return nil, fmt.Errorf("parse msp: %w", err)
		}

		endpointsCV, ok := ordererGroup.Groups[groupName].Values[channelconfig.EndpointsKey]
		var endpoints []string
		if ok {
			endpoints, err = ParseOrdererEndpoints(endpointsCV.Value)
			if err != nil {
				return nil, fmt.Errorf("parse endpoints: %w", err)
			}
		}

		orderersCfg[groupName] = &block.OrdererConfig{
			Name:      groupName,
			Msp:       mspCfg,
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

func ParseAnchorPeers(mspConfigGroup *common.ConfigGroup) ([]*peer.AnchorPeer, error) {
	if cv, ok := mspConfigGroup.Values[channelconfig.AnchorPeersKey]; ok {
		return ParseAnchorPeersFromBytes(cv.Value)
	}
	return []*peer.AnchorPeer{}, nil
}

func ParseAnchorPeersFromBytes(b []byte) ([]*peer.AnchorPeer, error) {
	anchorPeers := &peer.AnchorPeers{}
	if err := proto.Unmarshal(b, anchorPeers); err != nil {
		return nil, fmt.Errorf("unmarshal anchor peers: %w", err)
	}
	return anchorPeers.AnchorPeers, nil
}

func ParseOrdererBatchSize(cfg common.Config) (*orderer.BatchSize, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.OrdererGroupKey)
	}

	batchSize, exists := ordererGroup.Values[channelconfig.BatchSizeKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.BatchSizeKey)
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

func ParseOrdererBatchTimeout(cfg common.Config) (string, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", channelconfig.OrdererGroupKey)
	}

	batchTimeout, exists := ordererGroup.Values[channelconfig.BatchTimeoutKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", channelconfig.BatchTimeoutKey)
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

func ParseOrdererConsensusType(cfg common.Config) (*orderer.ConsensusType, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.OrdererGroupKey)
	}

	consensusType, exists := ordererGroup.Values[channelconfig.ConsensusTypeKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.ConsensusTypeKey)
	}

	return ParseOrdererConsensusTypeFromBytes(consensusType.Value)
}

func ParseOrdererConsensusTypeFromBytes(b []byte) (*orderer.ConsensusType, error) {
	ct := &orderer.ConsensusType{}
	if err := proto.Unmarshal(b, ct); err != nil {
		return nil, fmt.Errorf("unmarshal ConsensusType: %w", err)
	}
	return ct, nil
}

func ParseConsortium(cfg common.Config) (string, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", channelconfig.OrdererGroupKey)
	}

	consensusType, exists := ordererGroup.Values[channelconfig.ConsensusTypeKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", channelconfig.ConsensusTypeKey)
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

func ParseHashingAlgorithm(cfg common.Config) (string, error) {
	hashingAlgorithm, exists := cfg.ChannelGroup.Values[channelconfig.HashingAlgorithmKey]
	if !exists {
		return "", fmt.Errorf("%v type group doesn't exists", channelconfig.HashingAlgorithmKey)
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

func ParseBlockDataHashingStructure(cfg common.Config) (*common.BlockDataHashingStructure, error) {
	bdh, exists := cfg.ChannelGroup.Values[channelconfig.BlockDataHashingStructureKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.BlockDataHashingStructureKey)
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

func ParsePolicy(policiesCfg map[string]*common.ConfigPolicy) (map[string]*block.Policy, error) {
	policies := make(map[string]*block.Policy)

	for policyKey, policyCfg := range policiesCfg {
		switch policyCfg.Policy.Type {
		case int32(common.Policy_UNKNOWN):
			return policies, nil

		case int32(common.Policy_MSP):
			// never have seen this type
			// todo if you'll find it
			return policies, nil

		case int32(common.Policy_IMPLICIT_META):
			implicitMetaPolicy := &common.ImplicitMetaPolicy{}
			if err := proto.Unmarshal(policyCfg.Policy.Value, implicitMetaPolicy); err != nil {
				return nil, fmt.Errorf("unmarshal implicit meta policy from config: %w", err)
			}

			policies[policyKey] = &block.Policy{
				Policy: &block.Policy_Implicit{
					Implicit: implicitMetaPolicy,
				},
			}

		case int32(common.Policy_SIGNATURE):
			signaturePolicyEnv := &common.SignaturePolicyEnvelope{}
			if err := proto.Unmarshal(policyCfg.Policy.Value, signaturePolicyEnv); err != nil {
				return nil, fmt.Errorf("unmarshal implicit meta policy from config: %w", err)
			}

			policies[policyKey] = &block.Policy{
				Policy: &block.Policy_SignaturePolicy{
					SignaturePolicy: signaturePolicyEnv,
				},
			}
		}
	}

	return policies, nil
}
