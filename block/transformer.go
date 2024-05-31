package block

import (
	"github.com/s7techlab/hlf-sdk-go/proto/block"
)

// Transformer transforms parsed observer data. For example decrypt, or transformer protobuf state to json
type Transformer interface {
	Transform(*block.Block) (*block.Block, error)
}
