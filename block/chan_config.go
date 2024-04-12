package block

import (
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/channelconfig"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ErrUnknownFabricVersion = errors.New(`unknown fabric version`)
)

type FabricVersion string

const (
	FabricVersionUndefined FabricVersion = "undefined"
	FabricV1               FabricVersion = "1"
	FabricV2               FabricVersion = "2"
)

func FabricVersionIsV2(isV2 bool) FabricVersion {
	if isV2 {
		return FabricV2
	}

	return FabricV1
}

func (x *ChannelConfig) ToJSON() ([]byte, error) {
	opt := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	return opt.Marshal(x)
}

func UnmarshalChannelConfig(b []byte) (*ChannelConfig, error) {
	cc := &ChannelConfig{}

	if err := protojson.Unmarshal(b, cc); err != nil {
		return nil, err
	}

	return cc, nil
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

func ParseApplicationConfig(cfg common.Config) (map[string]*ApplicationConfig, error) {
	applicationGroup, exists := cfg.ChannelGroup.Groups[channelconfig.ApplicationGroupKey]
	if !exists {
		return nil, fmt.Errorf("application group doesn't exists")
	}

	appCfg := map[string]*ApplicationConfig{}

	for groupName := range applicationGroup.Groups {
		var (
			mspCfg   *MSP
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

		appCfg[groupName] = &ApplicationConfig{
			Name:        groupName,
			Msp:         mspCfg,
			AnchorPeers: ancPeers,
		}
	}

	return appCfg, nil
}

func ParseMSP(mspConfigGroup *common.ConfigGroup, groupName string) (*MSP, error) {
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

	return &MSP{Config: fabricMSPCfg, Policy: policy, Name: groupName}, nil
}

func ParseOrderer(cfg common.Config) (map[string]*OrdererConfig, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return nil, nil
	}
	orderersCfg := map[string]*OrdererConfig{}

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

		orderersCfg[groupName] = &OrdererConfig{
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

func ParsePolicy(policiesCfg map[string]*common.ConfigPolicy) (map[string]*Policy, error) {
	policies := make(map[string]*Policy)

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

			policies[policyKey] = &Policy{
				Policy: &Policy_Implicit{
					Implicit: implicitMetaPolicy,
				},
			}

		case int32(common.Policy_SIGNATURE):
			signaturePolicyEnv := &common.SignaturePolicyEnvelope{}
			if err := proto.Unmarshal(policyCfg.Policy.Value, signaturePolicyEnv); err != nil {
				return nil, fmt.Errorf("unmarshal implicit meta policy from config: %w", err)
			}

			policies[policyKey] = &Policy{
				Policy: &Policy_SignaturePolicy{
					SignaturePolicy: signaturePolicyEnv,
				},
			}
		}
	}

	return policies, nil
}

/* structs methods */

// GetAllCertificates - returns all(root, intermediate, admins) certificates from all MSPs'
func (x *ChannelConfig) GetAllCertificates() ([]*Certificate, error) {
	var certs []*Certificate

	for mspID := range x.Applications {
		cs, err := x.Applications[mspID].Msp.GetAllCertificates()
		if err != nil {
			return nil, fmt.Errorf("get all msps certificates: %w", err)
		}
		certs = append(certs, cs...)
	}

	for mspID := range x.Orderers {
		cs, err := x.Orderers[mspID].Msp.GetAllCertificates()
		if err != nil {
			return nil, fmt.Errorf("get all orderers certificates: %w", err)
		}
		certs = append(certs, cs...)
	}

	return certs, nil
}

func (x *ChannelConfig) FabricVersion() FabricVersion {
	if x.Capabilities != nil {
		_, isFabricV2 := x.Capabilities.Capabilities["V2_0"]
		if isFabricV2 {
			return FabricV2
		}
		return FabricV1
	}
	return FabricVersionUndefined
}

// GetAllCertificates - returns all certificates from MSP
func (x *MSP) GetAllCertificates() ([]*Certificate, error) {
	var certs []*Certificate

	for i := range x.Config.RootCerts {
		cert, err := NewCertificate(x.Config.RootCerts[i], CertType_CERT_TYPE_CA, x.Config.Name, x.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	for i := range x.Config.IntermediateCerts {
		cert, err := NewCertificate(x.Config.IntermediateCerts[i], CertType_CERT_TYPE_INTERMEDIATE, x.Config.Name, x.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	for i := range x.Config.Admins {
		cert, err := NewCertificate(x.Config.Admins[i], CertType_CERT_TYPE_ADMIN, x.Config.Name, x.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

func NewCertificate(cert []byte, t CertType, mspID, mspName string) (*Certificate, error) {
	b, _ := pem.Decode(cert)
	if b == nil {
		return &Certificate{}, fmt.Errorf("decode %s cert of %s", t, mspID)
	}

	c := &Certificate{
		Data:    cert,
		MspId:   mspID,
		Type:    t,
		MspName: mspName,
	}
	c.setCertificateSHA256(b)

	return c, nil
}

func (x *Certificate) setCertificateSHA256(b *pem.Block) {
	f := CalcCertificateSHA256(b)
	x.Fingerprint = f[:]
}

func CalcCertificateSHA256(b *pem.Block) []byte {
	f := sha256.Sum256(b.Bytes)
	return f[:]
}
