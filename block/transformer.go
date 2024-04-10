package block

// Transformer transforms parsed observer data. For example decrypt, or transformer protobuf state to json
type Transformer interface {
	Transform(*Block) (*Block, error)
}
