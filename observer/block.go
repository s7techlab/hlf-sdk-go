package observer

import (
	"github.com/hyperledger/fabric-protos-go/common"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type Block[T any] struct {
	Channel string
	Block   T
}

type CommonBlock struct {
	*Block[*common.Block]
}

type ParsedBlock struct {
	*Block[*hlfproto.Block]
}
