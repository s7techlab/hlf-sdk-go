package system

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/msp"
	"github.com/s7techlab/hlf-sdk-go/api"
	peerSDK "github.com/s7techlab/hlf-sdk-go/peer"
)

type csccV2 struct {
	*csccV1
}

// These are function names from Invoke first parameter
const (
	GetChannelConfig string = "GetChannelConfig"
)

func (c csccV2) GetChannelConfig(ctx context.Context, channelName string) (*common.Config, error) {
	resp, err := c.endorse(ctx, GetChannelConfig, channelName)
	if err != nil {
		return nil, err
	}
	var chConfig common.Config
	if err = proto.Unmarshal(resp, &chConfig); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	return &chConfig, nil
}

func NewCSCCV2(peerPool api.PeerPool, identity msp.SigningIdentity) api.CSCC {
	return &csccV2{
		csccV1: &csccV1{
			peerPool:  peerPool,
			identity:  identity,
			processor: peerSDK.NewProcessor(``),
		},
	}
}
