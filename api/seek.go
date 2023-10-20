package api

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

// NewSeekSpecified returns orderer.SeekPosition_Specified position
func NewSeekSpecified(number uint64) *orderer.SeekPosition {
	return &orderer.SeekPosition{Type: &orderer.SeekPosition_Specified{Specified: &orderer.SeekSpecified{Number: number}}}
}

func NewSeekInfo(start, stop *orderer.SeekPosition) *orderer.SeekInfo {
	return &orderer.SeekInfo{
		Start:    start,
		Stop:     stop,
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}
}

func NewMarshalledSeekInfo(start, stop *orderer.SeekPosition) ([]byte, error) {
	return proto.Marshal(NewSeekInfo(start, stop))
}
