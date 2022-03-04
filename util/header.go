package util

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
)

// NewPayloadFromHeader returns marshalled payload from headers and data
func NewPayloadFromHeader(channelHeader, signatureHeader, data []byte) ([]byte, error) {
	return proto.Marshal(&common.Payload{Header: &common.Header{ChannelHeader: channelHeader, SignatureHeader: signatureHeader}, Data: data})
}

// NewChannelHeader returns new channel header bytes for presented transaction and channel
func NewChannelHeader(headerType common.HeaderType, txId string, channelId string, epoch uint64, extension *peer.ChaincodeHeaderExtension) ([]byte, error) {
	ts, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}
	var channelName string

	if len(channelId) > 0 {
		channelName = channelId
	}
	payloadChannelHeader := &common.ChannelHeader{
		Type:      int32(headerType),
		Version:   1,
		Timestamp: ts,
		ChannelId: channelName,
		Epoch:     epoch,
		TxId:      txId,
	}

	if extension != nil {
		serExt, err := proto.Marshal(extension)
		if err != nil {
			return nil, err
		}
		payloadChannelHeader.Extension = serExt
	}
	return proto.Marshal(payloadChannelHeader)
}

// NewSignatureHeader returns marshalled signature header for presented identity
func NewSignatureHeader(id msp.SigningIdentity, nonce []byte) ([]byte, error) {
	sh := new(common.SignatureHeader)
	if creator, err := id.Serialize(); err != nil {
		return nil, errors.Wrap(err, `failed to serialize identity`)
	} else {
		sh.Creator = creator
	}
	sh.Nonce = nonce
	shBytes, err := proto.Marshal(sh)
	if err != nil {
		return nil, err
	}
	return shBytes, nil
}
