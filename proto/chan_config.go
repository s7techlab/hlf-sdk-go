package proto

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/channelconfig"
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

	Policy map[PolicyKey]Policy `json:"policy"`
}

type ApplicationConfig struct {
	Name        string             `json:"name"`
	MSP         MSP                `json:"msp"`
	AnchorPeers []*peer.AnchorPeer `json:"anchor_peers"`
}

type MSP struct {
	Config msp.FabricMSPConfig  `json:"config"`
	Policy map[PolicyKey]Policy `json:"policy"`
}

type OrdererConfig struct {
	Name      string   `json:"name"`
	MSP       MSP      `json:"msp"`
	Endpoints []string `json:"endpoints"`
}

type PolicyKey string

const (
	ReadersPolicyKey PolicyKey = "Readers"
	WritersPolicyKey PolicyKey = "Writers"
	AdminsPolicyKey  PolicyKey = "Admins"

	LifecycleEndporsementPolicyKey PolicyKey = "LifecycleEndporsement"
	EndporsementPolicyKey          PolicyKey = "Endporsement"
)

// Policy - can contain different policies: implicit, signature
// check type and for 'nil' before usage
type Policy struct {
	Type               common.Policy_PolicyType        `json:"-"`
	ImplicitMetaPolicy *common.ImplicitMetaPolicy      `json:"implicit,omitempty"`
	SignaturePolicy    *common.SignaturePolicyEnvelope `json:"signature,omitempty"`
}

// Certificate - describes certificate(can be ca, intermediate, admin) from msp
type Certificate struct {
	FingerprintSHA256 []byte
	Data              []byte
	MSPID             string
	Type              CertType
	MSPName           string
}

type CertType string

const (
	RootCACertType       CertType = "ca"
	IntermediateCertType CertType = "intermediate"
	AdminCertType        CertType = "admin"
)

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

	policies, err := ParsePolicy(cc.ChannelGroup.Policies)
	if err != nil {
		return nil, fmt.Errorf("parse policies: %w", err)
	}

	chanCfg.Policy = policies

	return chanCfg, nil
}

func ParseApplicationConfig(cfg common.Config) (map[string]ApplicationConfig, error) {
	applicationGroup, exists := cfg.ChannelGroup.Groups[channelconfig.ApplicationGroupKey]
	if !exists {
		return nil, fmt.Errorf("application group doesn't exists")
	}

	appCfg := map[string]ApplicationConfig{}

	for groupName := range applicationGroup.Groups {
		mspCfg, err := ParseMSP(applicationGroup.Groups[groupName])
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

func ParseMSP(mspConfigGroup *common.ConfigGroup) (*MSP, error) {
	mspCV, ok := mspConfigGroup.Values[channelconfig.MSPKey]
	if !ok {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.MSPKey)
	}

	mspCfg := &msp.MSPConfig{}
	if err := proto.Unmarshal(mspCV.Value, mspCfg); err != nil {
		return nil, fmt.Errorf("unmarshal MSPConfig: %w", err)
	}

	fmspCfg := &msp.FabricMSPConfig{}
	if err := proto.Unmarshal(mspCfg.Config, fmspCfg); err != nil {
		return nil, fmt.Errorf("unmarshal FabricMSPConfig: %w", err)
	}

	policy, err := ParsePolicy(mspConfigGroup.Policies)
	if err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}

	return &MSP{Config: *fmspCfg, Policy: policy}, nil
}

func ParseAnchorPeers(b []byte) ([]*peer.AnchorPeer, error) {
	anchorPeers := &peer.AnchorPeers{}
	if err := proto.Unmarshal(b, anchorPeers); err != nil {
		return nil, fmt.Errorf("unmarshal anchor peers: %w", err)
	}

	return anchorPeers.AnchorPeers, nil
}

func ParseOrderer(cfg common.Config) (map[string]OrdererConfig, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.OrdererGroupKey)
	}
	orderersCfg := map[string]OrdererConfig{}

	for groupName := range ordererGroup.Groups {
		mspCfg, err := ParseMSP(ordererGroup.Groups[groupName])
		if err != nil {
			return nil, fmt.Errorf("parse msp: %w", err)
		}

		endpointsCV, ok := ordererGroup.Groups[groupName].Values[channelconfig.EndpointsKey]
		if !ok {
			return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.EndpointsKey)
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

//
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

//
func ParseOrdererConsesusType(cfg common.Config) (*orderer.ConsensusType, error) {
	ordererGroup, exists := cfg.ChannelGroup.Groups[channelconfig.OrdererGroupKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.OrdererGroupKey)
	}

	consensusType, exists := ordererGroup.Values[channelconfig.ConsensusTypeKey]
	if !exists {
		return nil, fmt.Errorf("%v type group doesn't exists", channelconfig.ConsensusTypeKey)
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

//
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

//
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

//
func ParsePolicy(policiesCfg map[string]*common.ConfigPolicy) (map[PolicyKey]Policy, error) {
	policies := make(map[PolicyKey]Policy)

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

			policies[PolicyKey(policyKey)] = Policy{
				Type:               common.Policy_PolicyType(policyCfg.Policy.Type),
				ImplicitMetaPolicy: implicitMetaPolicy,
			}

		case int32(common.Policy_SIGNATURE):
			signaturePolicyEnv := &common.SignaturePolicyEnvelope{}
			if err := proto.Unmarshal(policyCfg.Policy.Value, signaturePolicyEnv); err != nil {
				return nil, fmt.Errorf("unmarshal implicit meta policy from config: %w", err)
			}

			policies[PolicyKey(policyKey)] = Policy{
				Type:            common.Policy_PolicyType(policyCfg.Policy.Type),
				SignaturePolicy: signaturePolicyEnv,
			}
		}
	}

	return policies, nil
}

/* structs methods */
// GetAllCertificates - returns all(root, intermediate, admins) certificates from all MSP's
func (c ChannelConfig) GetAllCertificates() ([]Certificate, error) {
	var certs []Certificate

	for mspID := range c.Applications {
		cs, err := c.Applications[mspID].MSP.GetAllCertificates()
		if err != nil {
			return nil, fmt.Errorf("get all msps certificates: %w", err)
		}
		certs = append(certs, cs...)
	}

	for mspID := range c.Orderers {
		cs, err := c.Applications[mspID].MSP.GetAllCertificates()
		if err != nil {
			return nil, fmt.Errorf("get all orderers certificates: %w", err)
		}
		certs = append(certs, cs...)
	}

	return certs, nil
}

// GetAllCertificates - returns all certificates from MSP
func (c MSP) GetAllCertificates() ([]Certificate, error) {
	var certs []Certificate

	for i := range c.Config.RootCerts {
		cert, err := NewCertificate(c.Config.Admins[i], RootCACertType, c.Config.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	for i := range c.Config.IntermediateCerts {
		cert, err := NewCertificate(c.Config.Admins[i], IntermediateCertType, c.Config.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	for i := range c.Config.Admins {
		cert, err := NewCertificate(c.Config.Admins[i], AdminCertType, c.Config.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

func NewCertificate(cert []byte, t CertType, mspID string) (Certificate, error) {
	b, _ := pem.Decode(cert)
	if b == nil {
		return Certificate{}, fmt.Errorf("decode %s cert of %s", t, mspID)
	}

	c := Certificate{
		Data:  cert,
		MSPID: mspID,
		Type:  t,
	}
	c.setCertificateSHA256(b)

	return c, nil
}

func (c *Certificate) setCertificateSHA256(b *pem.Block) {
	f := sha256.Sum256(b.Bytes)
	c.FingerprintSHA256 = f[:]
}
