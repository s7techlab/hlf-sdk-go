package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func NewCommonHeader(headerType common.HeaderType, txID string, nonce []byte, ts *timestamp.Timestamp, serializedCreator []byte, channel, chaincode string) (*common.Header, error) {
	var extension *peer.ChaincodeHeaderExtension
	if chaincode != `` {
		extension = &peer.ChaincodeHeaderExtension{ChaincodeId: &peer.ChaincodeID{Name: chaincode}}
	}

	channelHeader, err := NewMarshaledChannelHeader(headerType, txID, ts, channel, 0, extension)
	if err != nil {
		return nil, fmt.Errorf(`channel header: %w`, err)
	}

	signatureHeader, err := NewMarshalledSignatureHeader(serializedCreator, nonce)
	if err != nil {
		return nil, fmt.Errorf(`signature header: %w`, err)
	}

	return &common.Header{
		ChannelHeader:   channelHeader,
		SignatureHeader: signatureHeader,
	}, nil
}

func NewMarshalledCommonHeader(headerType common.HeaderType, txID string, nonce []byte, ts *timestamp.Timestamp, serializedCreator []byte, channel, chaincode string) ([]byte, error) {
	header, err := NewCommonHeader(headerType, txID, nonce, ts, serializedCreator, channel, chaincode)
	if err != nil {
		return nil, fmt.Errorf(`create common header: %w`, err)
	}
	return proto.Marshal(header)
}

// NewMarshaledChannelHeader returns new channel header bytes for presented transaction and channel
func NewMarshaledChannelHeader(headerType common.HeaderType, txId string, ts *timestamp.Timestamp, channelId string, epoch uint64, extension *peer.ChaincodeHeaderExtension) ([]byte, error) {
	var channelName string

	if len(channelId) > 0 {
		channelName = channelId
	}
	channelHeader := &common.ChannelHeader{
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
			return nil, fmt.Errorf(`channel header extension: %w`, err)
		}
		channelHeader.Extension = serExt
	}
	return proto.Marshal(channelHeader)
}

// NewMarshalledSignatureHeader returns marshalled signature header for presented identity
func NewMarshalledSignatureHeader(serializedCreator []byte, nonce []byte) ([]byte, error) {
	sh := &common.SignatureHeader{
		Creator: serializedCreator,
		Nonce:   nonce,
	}
	return proto.Marshal(sh)
}
