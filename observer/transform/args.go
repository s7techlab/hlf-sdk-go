package transform

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

type (
	InputArgsTransformer interface {
		Transform([][]byte) error
	}
	InputArgsMatch  func([][]byte) bool
	InputArgsMutate func([][]byte) error

	InputArgs struct {
		match  InputArgsMatch
		mutate InputArgsMutate
	}
)

func NewInputArgs(match InputArgsMatch, mutate InputArgsMutate) *InputArgs {
	return &InputArgs{
		match:  match,
		mutate: mutate,
	}
}

func (tr *InputArgs) Transform(args [][]byte) error {
	if tr.match(args) {
		return tr.mutate(args)
	}
	return nil
}

func InputArgsProto(fn string, target proto.Message) *InputArgs {
	return NewInputArgs(
		InputArgsMatchFunc(fn),
		InputArgsMutateProto(target),
	)
}

func InputArgsMatchFunc(fn string) InputArgsMatch {
	return InputArgsMatchString(0, fn) // fn is on pos=0
}

func InputArgsMatchString(pos int, str string) InputArgsMatch {
	return func(args [][]byte) bool {
		if len(args) > pos && string(args[pos]) == str {
			return true
		}
		return false
	}
}

func InputArgsMutateProtoAtPos(target proto.Message, pos int) InputArgsMutate {
	return func(args [][]byte) error {
		if len(args) < pos+1 || len(args[pos]) == 0 {
			return nil
		}
		arg, err := Proto2JSON(args[pos], target)
		if err != nil {
			return fmt.Errorf(`args mutator pos=%d: %w`, pos, err)
		}

		args[pos] = arg
		return nil
	}
}

func InputArgsMutateProto(target proto.Message) InputArgsMutate {
	return InputArgsMutateProtoAtPos(target, 1) // args[0] - func name, args[1] - proto by default
}
