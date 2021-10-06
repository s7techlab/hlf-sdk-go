package auto

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
)

func getOrdererAddresses(cfg *common.Config) ([]string, error) {
	if cfg.ChannelGroup == nil {
		// TODO what to return?
		return []string{}, nil
	}

	if _, ok := cfg.ChannelGroup.Values["OrdererAddresses"]; !ok {
		return []string{}, nil
	}

	ordererAddresses := cfg.ChannelGroup.Values["OrdererAddresses"].Value
	od := &common.OrdererAddresses{}
	if err := proto.Unmarshal(ordererAddresses, od); err != nil {
		return nil, err
	}

	return od.Addresses, nil
}
