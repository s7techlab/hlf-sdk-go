package transform

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jsonpbMarshaler = &jsonpb.Marshaler{EmitDefaults: true}

func Proto2JSON(serialized []byte, target proto.Message) ([]byte, error) {
	m := proto.Clone(target)

	if err := proto.Unmarshal(serialized, m); err != nil {
		return nil, fmt.Errorf(`proto unmarshal to=%s: %w`, reflect.TypeOf(target), err)
	}

	s, err := jsonpbMarshaler.MarshalToString(m)
	if err != nil {
		return nil, fmt.Errorf(`json pb marshal: %w`, err)
	}

	return []byte(s), nil
}
