package util

import (
	"bytes"
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
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

func ProceedChannelUpdate(ctx context.Context, channelName string, update *common.ConfigUpdate, orderer api.Orderer, id msp.SigningIdentity) error {
	confUpdBytes, err := proto.Marshal(update)
	if err != nil {
		return errors.Wrap(err, `failed to marshal common.ConfigUpdate`)
	}

	txId, nonce, err := NewTxWithNonce(id)
	if err != nil {
		return errors.Wrap(err, `failed to get nonce`)
	}

	signatureHeader, err := NewSignatureHeader(id, nonce)
	if err != nil {
		return errors.Wrap(err, `failed to get signature header`)
	}

	buf := bytes.NewBuffer(signatureHeader)
	buf.Write(confUpdBytes)

	signature, err := id.Sign(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, `failed to sign bytes`)
	}

	sig := &common.ConfigSignature{
		SignatureHeader: signatureHeader,
		Signature:       signature,
	}

	confUpdEnvelope := &common.ConfigUpdateEnvelope{
		ConfigUpdate: confUpdBytes,
		Signatures:   []*common.ConfigSignature{sig},
	}

	confUpdEnvBytes, err := proto.Marshal(confUpdEnvelope)
	if err != nil {
		return errors.Wrap(err, `failed to marshal common.ConfigUpdateEnvelope`)
	}

	channelHeader, err := NewChannelHeader(common.HeaderType_CONFIG_UPDATE, txId, channelName, 0, nil)
	if err != nil {
		return errors.Wrap(err, `failed to get channel header`)
	}

	payload, err := NewPayloadFromHeader(channelHeader, signatureHeader, confUpdEnvBytes)
	if err != nil {
		return errors.Wrap(err, `failed to get payload`)
	}

	envelope := &common.Envelope{
		Payload: payload,
	}

	envelope.Signature, err = id.Sign(envelope.Payload)
	if err != nil {
		return errors.WithMessage(err, "signing payload failed")
	}

	if _, err := orderer.Broadcast(ctx, envelope); err != nil {
		return errors.WithMessage(err, "failed broadcast to orderer")
	}

	return nil
}
