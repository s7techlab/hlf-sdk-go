package observer

// BlockTransformer transforms parsed observer data. For example decrypt, or transformer protobuf state to json
type BlockTransformer interface {
	Transform(*ParsedBlock) error
}
