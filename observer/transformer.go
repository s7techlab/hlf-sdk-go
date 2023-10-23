package observer

// BlockTransformer transforms parsed Observer data. For example decrypt, or transformer protobuf state to json
type BlockTransformer interface {
	Transform(*ParsedBlock) error
}
