package tx

import (
	"context"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"

	"github.com/atomyze-ru/hlf-sdk-go/api"
)

func InvokeProto(ctx context.Context, invoker api.Invoker, channel, chaincode string, args []interface{}, target proto.Message) (proto.Message, error) {
	argsBytes, err := ArgsBytes(args...)
	if err != nil {
		return nil, err
	}

	return InvokeBytesProto(ctx, invoker, channel, chaincode, argsBytes, target)
}

func InvokeStringsProto(ctx context.Context, invoker api.Invoker, channel, chaincode string, args []string, target proto.Message) (proto.Message, error) {
	return InvokeBytesProto(ctx, invoker, channel, chaincode, StringArgsBytes(args...), target)
}

func InvokeBytesProto(ctx context.Context, invoker api.Invoker, channel, chaincode string, args [][]byte, target proto.Message) (proto.Message, error) {
	res, _, err := invoker.Invoke(
		ctx, channel, chaincode, args, nil, nil, ``)

	if err != nil {
		return nil, fmt.Errorf(`invoke channel=%s chaincode=%s: %w`, channel, chaincode, err)
	}

	resProto := proto.Clone(target)

	if err = proto.Unmarshal(res.Payload, resProto); err != nil {
		return nil, fmt.Errorf(`unmarshal invoke result to %s: %w`, reflect.TypeOf(target), err)
	}

	return resProto, nil
}
