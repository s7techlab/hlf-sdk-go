package tx

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
)

var (
	ErrUnknownArgType = errors.New(`unknown arg type`)
)

func FnArgs(fn string, args ...[]byte) [][]byte {
	var argsBytes [][]byte

	if fn != `` {
		argsBytes = append(argsBytes, []byte(fn))
	}

	return append(argsBytes, args...)
}

func ArgBytes(arg interface{}) ([]byte, error) {
	switch val := arg.(type) {

	case []byte:
		return val, nil

	case string:
		return []byte(val), nil

	case proto.Message:
		res, err := proto.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf(`marshal proto arg: %w`, err)
		}

		return res, nil

	default:
		return nil, ErrUnknownArgType
	}
}

func StringArgsBytes(args ...string) [][]byte {
	var argsBytes [][]byte

	for _, arg := range args {
		argsBytes = append(argsBytes, []byte(arg))
	}

	return argsBytes
}

func ArgsBytes(args ...interface{}) ([][]byte, error) {
	var argsBytes [][]byte

	for pos, arg := range args {

		converted, err := ArgBytes(arg)
		if err != nil {
			return nil, fmt.Errorf(`args[%d]: %w`, pos, err)
		}

		argsBytes = append(argsBytes, converted)
	}

	return argsBytes, nil
}
