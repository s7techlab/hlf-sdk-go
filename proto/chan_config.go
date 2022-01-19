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
		return nil, fmt.Errorf("orderer group doesn't exists")
	}
	orderersCfg := map[string]OrdererConfig{}

	for groupName := range ordererGroup.Groups {
		mspCfg, err := ParseMSP(ordererGroup.Groups[groupName].Values[mspKey].Value)
		if err != nil {
			return nil, fmt.Errorf("parse msp: %w", err)
		}

		endpoints, err := ParseEndpoints(ordererGroup.Groups[groupName].Values[endpointsKey].Value)
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

func ParseEndpoints(b []byte) ([]string, error) {
	oa := &common.OrdererAddresses{}
	if err := proto.Unmarshal(b, oa); err != nil {
		return nil, fmt.Errorf("unmarshal OrdererAddresses: %w", err)
	}
	return oa.Addresses, nil
}

// func ParseChannelConfig(cfg *common.Config) (cc ChannelConfig, cs []Certificate, err error) {

// 	os := map[string]OrdererConfig{}

// 	for gName, g := range ordererGroup.Groups {
// 		var (
// 			o     OrdererConfig
// 			mspCs []Certificate
// 		)

// 		o.Name = gName

// 		var msp *common.ConfigValue

// 		msp, exists = g.Values[mspKey]
// 		if exists {
// 			o.MSP, mspCs, err = parseMSP(msp.Value)
// 			if err != nil {
// 				return cc, cs, fmt.Errorf("parse MSP: %w", err)
// 			}

// 			for i := range mspCs {
// 				mspCs[i].MSPName = gName
// 			}

// 			cs = append(cs, mspCs...)
// 		}

// 		var endpoints *common.ConfigValue

// 		endpoints, exists = g.Values[endpointsKey]
// 		if exists {
// 			var as common.OrdererAddresses
// 			if err := proto.Unmarshal(endpoints.Value, &as); err != nil {
// 				// TODO: log error
// 				//logrus.WithError(err).Error(
// 				//	"failed to proto unmarshal OrdererEndpointConfig")
// 			} else {
// 				o.Endpoints = as.Addresses
// 			}
// 		}

// 		os[gName] = o
// 	}

// 	batchSize, exists := ordererGroup.Values[batchSizeKey]
// 	if exists {
// 		var bs orderer.BatchSize
// 		err := jsonpb.Unmarshal(bytes.NewReader(batchSize.Value), &bs)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to JSON unmarshal BatchSize")
// 		} else {
// 			cc.OrdererBatchSize = bs
// 		}
// 	}

// 	batchTimeout, exists := ordererGroup.Values[batchTimeoutKey]
// 	if exists {
// 		var bt orderer.BatchTimeout
// 		err := jsonpb.Unmarshal(bytes.NewReader(batchTimeout.Value), &bt)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to JSON unmarshal BatchTimeout")
// 		} else {
// 			cc.OrdererBatchTimeout = bt.Timeout
// 		}
// 	}

// 	consensusType, exists := ordererGroup.Values[consensusTypeKey]
// 	if exists {
// 		var ct orderer.ConsensusType
// 		err := jsonpb.Unmarshal(bytes.NewReader(consensusType.Value), &ct)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to JSON unmarshal ConsensusType")
// 		} else {
// 			cc.OrdererConsesusType = ct
// 		}
// 	}

// 	consortium, exists := cfg.ChannelGroup.Values[consortiumKey]
// 	if exists {
// 		var c common.Consortium
// 		err := proto.Unmarshal(consortium.Value, &c)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to proto unmarshal Consortium")
// 		} else {
// 			cc.Consortium = c.Name
// 		}
// 	}

// 	hashingAlgorithm, exists := cfg.ChannelGroup.Values[hashingAlgorithmKey]
// 	if exists {
// 		var ha common.HashingAlgorithm
// 		err := proto.Unmarshal(hashingAlgorithm.Value, &ha)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to proto unmarshal HashingAlgorithm")
// 		} else {
// 			cc.HashingAlgorithm = ha.Name
// 		}
// 	}

// 	ordererAddresses, exists := cfg.ChannelGroup.Values[ordererAddressesKey]
// 	if exists {
// 		var oas common.OrdererAddresses
// 		err := proto.Unmarshal(ordererAddresses.Value, &oas)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to proto unmarshal OrdererAddresses")
// 		} else {
// 			cc.OrdererAddresses = oas.Addresses
// 		}
// 	}

// 	blockDataHashingStructure, exists := cfg.ChannelGroup.
// 		Values[blockDataHashingStructureKey]
// 	if exists {
// 		var s common.BlockDataHashingStructure
// 		err := proto.Unmarshal(blockDataHashingStructure.Value, &s)
// 		if err != nil {
// 			// TODO: log error
// 			//logrus.WithError(err).Error(
// 			//	"failed to proto unmarshal BlockDataHashingStructure")
// 		} else {
// 			cc.BlockDataHashingStructure = s
// 		}
// 	}

// 	if len(os) > 0 {
// 		cc.Orderers = os
// 	}

// 	return cc, cs, nil
// }
