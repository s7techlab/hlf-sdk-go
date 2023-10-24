package observer

import (
	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type (
	ParsedBlock struct {
		Block         *hlfproto.Block // parsed block
		BlockOriginal *hlfproto.Block // here is original block before transformation if it is, otherwise it's nil
		Channel       string
		Error         error
	}
)
