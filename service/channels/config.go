package channels

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	mspproto "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/orderer/etcdraft"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/proto/channel"
)

var (
	ErrOrdererGroupNotFound = errors.New(`orderer addresses not found`)
	ErrEmptyChannelConfig   = errors.New(`empty channel config`)
)

func GetOrdererHostFromChannelConfig(conf *common.Config) (string, error) {
	ordValues, ok := conf.ChannelGroup.Values[channelconfig.OrdererAddressesKey]
	if !ok {
		return ``, ErrOrdererGroupNotFound
	}

	ordererAddresses := common.OrdererAddresses{}
	if err := proto.Unmarshal(ordValues.Value, &ordererAddresses); err != nil {
		return ``, errors.Wrap(err, `failed to unmarshal orderer addresses`)
	}

	// TODO return all addresses instead of first
	return ordererAddresses.Addresses[0], nil
}

// ProceedChannelUpdate - sends channel update config with signatures of all provided identities
func ProceedChannelUpdate(
	ctx context.Context,
	channelName string,
	update *common.ConfigUpdate,
	orderer api.Orderer,
	ids []msp.SigningIdentity,
) error {
	if len(ids) == 0 {
		return errors.New("no signing identities provided")
	}

	confUpdBytes, err := proto.Marshal(update)
	if err != nil {
		return errors.Wrap(err, `failed to marshal common.ConfigUpdate`)
	}

	serialized, err := ids[0].Serialize()
	if err != nil {
		return fmt.Errorf(`serialize identity: %w`, err)
	}

	txParams, err := tx.GenerateParamsForSerializedIdentity(serialized)
	if err != nil {
		return errors.Wrap(err, `tx id`)
	}

	signatures := make([]*common.ConfigSignature, len(ids))
	for i := range ids {
		signatures[i], err = signConfig(ids[i], confUpdBytes, txParams.Nonce)
		if err != nil {
			return errors.Wrap(err, `failed to sign config update`)
		}
	}

	confUpdEnvelope := &common.ConfigUpdateEnvelope{
		ConfigUpdate: confUpdBytes,
		Signatures:   signatures,
	}

	confUpdEnvBytes, err := proto.Marshal(confUpdEnvelope)
	if err != nil {
		return errors.Wrap(err, `failed to marshal common.ConfigUpdateEnvelope`)
	}

	channelHeader, err := hlfproto.NewCommonHeader(
		common.HeaderType_CONFIG_UPDATE,
		txParams.ID,
		txParams.Nonce,
		txParams.Timestamp,
		serialized,
		channelName,
		``,
		nil)
	if err != nil {
		return errors.Wrap(err, `failed to get channel header`)
	}

	payload, err := hlfproto.NewMarshalledCommonPayload(channelHeader, confUpdEnvBytes)
	if err != nil {
		return errors.Wrap(err, `failed to get payload`)
	}

	envelope := &common.Envelope{
		Payload: payload,
	}

	envelope.Signature, err = ids[0].Sign(envelope.Payload)
	if err != nil {
		return errors.WithMessage(err, "signing payload failed")
	}

	if _, err := orderer.Broadcast(ctx, envelope); err != nil {
		return errors.WithMessage(err, "failed broadcast to orderer")
	}

	return nil
}

func signConfig(id msp.SigningIdentity, configUpdateBytes, nonce []byte) (*common.ConfigSignature, error) {
	serialized, err := id.Serialize()
	if err != nil {
		return nil, err
	}
	signatureHeader, err := hlfproto.NewMarshalledSignatureHeader(serialized, nonce)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get signature header`)
	}

	buf := bytes.NewBuffer(signatureHeader)
	buf.Write(configUpdateBytes)

	signature, err := id.Sign(buf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, `failed to sign bytes`)
	}

	sig := &common.ConfigSignature{
		SignatureHeader: signatureHeader,
		Signature:       signature,
	}
	return sig, nil
}

const etcdRaft = "etcdraft"

// ExtractChannelConfigMSP returns channel message populated from configuration block
// Channel.Name WON'T be set, you have to set it manually
func ExtractChannelConfigMSP(config *common.Config) (*channel.ChannelConfigMSP, error) {
	channelConfigMSP := &channel.ChannelConfigMSP{}
	if config.ChannelGroup == nil {
		return nil, ErrEmptyChannelConfig
	}
	app, hasApp := config.ChannelGroup.Groups[channelconfig.ApplicationGroupKey]
	if hasApp {
		for _, orgGroup := range app.Groups {
			org := &channel.OrganizationMSPConfig{}
			if peersGroup, hasPeers := orgGroup.Values[channelconfig.AnchorPeersKey]; hasPeers {
				anchorPeers := peer.AnchorPeers{}
				if err := proto.Unmarshal(peersGroup.Value, &anchorPeers); err != nil {
					return nil, fmt.Errorf("unmarshal anchor peers: %w", err)
				}
				org.AnchorPeers = &peer.AnchorPeers{}
				org.AnchorPeers.AnchorPeers = append(org.AnchorPeers.AnchorPeers, anchorPeers.AnchorPeers...)
			}

			if mspGroup, hasMSP := orgGroup.Values[channelconfig.MSPKey]; hasMSP {
				mspCfg := &mspproto.MSPConfig{}
				if err := proto.Unmarshal(mspGroup.Value, mspCfg); err != nil {
					return nil, fmt.Errorf("unmarshal MSP config: %w", err)
				}
				if mspCfg.Type != int32(msp.FABRIC) {
					return nil, fmt.Errorf(
						"unsupported MSP type: %v",
						msp.ProviderTypeToString(msp.ProviderType(mspCfg.Type)),
					)
				}
				fabricMSPCfg := &mspproto.FabricMSPConfig{}
				if err := proto.Unmarshal(mspCfg.Config, fabricMSPCfg); err != nil {
					return nil, fmt.Errorf("unmarshal MSP: %w", err)
				}
				org.Msp = fabricMSPCfg
			}

			channelConfigMSP.Organizations = append(channelConfigMSP.Organizations, org)
		}
	}

	if ord, hasOrd := config.ChannelGroup.Groups[channelconfig.OrdererGroupKey]; hasOrd {
		if bsGroup, ok := ord.Values[channelconfig.BatchSizeKey]; ok {
			bs := &orderer.BatchSize{}
			if err := proto.Unmarshal(bsGroup.Value, bs); err != nil {
				return nil, fmt.Errorf("unmarshall batch size: %w", err)
			}
			channelConfigMSP.BatchSize = bs
		}
		if btGroup, ok := ord.Values[channelconfig.BatchTimeoutKey]; ok {
			bt := &orderer.BatchTimeout{}
			if err := proto.Unmarshal(btGroup.Value, bt); err != nil {
				return nil, fmt.Errorf("unmarshal batch timeout: %w", err)
			}
			channelConfigMSP.BatchTimeout = bt.GetTimeout()
		}

		for _, ordererGroup := range ord.Groups {
			if endpointsB, ok := ordererGroup.Values[channelconfig.EndpointsKey]; ok {
				addresses := &common.OrdererAddresses{}
				if err := proto.Unmarshal(endpointsB.Value, addresses); err != nil {
					return nil, fmt.Errorf("unmarshal endpoints: %w", err)
				}
				channelConfigMSP.Endpoints = append(channelConfigMSP.Endpoints, addresses.Addresses...)
			}
		}

		if etcdGroup, ok := ord.Values[channelconfig.ConsensusTypeKey]; ok {
			consensusType := &orderer.ConsensusType{}
			if err := proto.Unmarshal(etcdGroup.Value, consensusType); err != nil {
				return nil, fmt.Errorf("unmarshal consensus type: %w", err)
			}
			if consensusType.Type == etcdRaft {
				raft := &etcdraft.ConfigMetadata{}
				if err := proto.Unmarshal(consensusType.Metadata, raft); err != nil {
					return nil, fmt.Errorf("unmarshal etcd config: %w", err)
				}
				channelConfigMSP.EtcdraftOptions = raft.Options
				channelConfigMSP.EtcdraftNodes = raft.Consenters
			}
		}
	}

	return channelConfigMSP, nil
}
