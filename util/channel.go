package util

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	hlfproto "github.com/s7techlab/hlf-sdk-go/proto"
)

var (
	ErrOrdererGroupNotFound = errors.New(`orderer addresses not found`)
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
