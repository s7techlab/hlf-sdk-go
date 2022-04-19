package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/s7techlab/hlf-sdk-go/api"
)

var (
	ErrUnknownArgType = errors.New(`unknown arg type`)
)

func FnArgs(fn string, args [][]byte) [][]byte {
	return append([][]byte{[]byte(fn)}, args...)
}

func ArgBytes(arg interface{}) ([]byte, error) {
	switch val := arg.(type) {

	case string:
		return []byte(val), nil
	case proto.Message:
		return proto.Marshal(val)
	default:
		return nil, ErrUnknownArgType
	}
}

func StringArgsBytes(args []string) [][]byte {
	var argsBytes [][]byte

	for _, arg := range args {
		argsBytes = append(argsBytes, []byte(arg))
	}

	return argsBytes
}

func ArgsBytes(args []interface{}) ([][]byte, error) {
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

type ChaincodeProtoQuerier struct {
	Querier   api.Querier
	Channel   string
	Chaincode string
}

func (c *ChaincodeProtoQuerier) Query(ctx context.Context, args ...interface{}) (*peer.Response, error) {
	argsBytes, err := ArgsBytes(args)
	if err != nil {
		return nil, err
	}
	return c.Querier.Query(ctx, c.Channel, c.Chaincode, argsBytes, nil, nil)
}

func (c *ChaincodeProtoQuerier) QueryBytes(ctx context.Context, args ...[]byte) (*peer.Response, error) {
	return c.Querier.Query(ctx, c.Channel, c.Chaincode, args, nil, nil)
}

func (c *ChaincodeProtoQuerier) QueryProto(ctx context.Context, args []interface{}, target proto.Message) (proto.Message, error) {
	return QueryProto(ctx, c.Querier, c.Channel, c.Chaincode, args, target)
}

func (c *ChaincodeProtoQuerier) QueryStringsProto(ctx context.Context, args []string, target proto.Message) (proto.Message, error) {
	return QueryBytesProto(ctx, c.Querier, c.Channel, c.Chaincode, StringArgsBytes(args), target)
}

func (c *ChaincodeProtoQuerier) QueryBytesProto(ctx context.Context, args [][]byte, target proto.Message) (proto.Message, error) {
	return QueryBytesProto(ctx, c.Querier, c.Channel, c.Chaincode, args, target)
}

func NewChaincodeProtoQuerier(querier api.Querier, channel, chaincode string) *ChaincodeProtoQuerier {
	return &ChaincodeProtoQuerier{
		Querier:   querier,
		Channel:   channel,
		Chaincode: chaincode,
	}
}

func QueryProto(ctx context.Context, querier api.Querier, channel, chaincode string, args []interface{}, target proto.Message) (proto.Message, error) {
	argsBytes, err := ArgsBytes(args)
	if err != nil {
		return nil, err
	}

	return QueryBytesProto(ctx, querier, channel, chaincode, argsBytes, target)
}

func QueryBytesProto(ctx context.Context, querier api.Querier, channel, chaincode string, args [][]byte, target proto.Message) (proto.Message, error) {

	res, err := querier.Query(
		ctx, channel, chaincode, args, nil, nil)

	if err != nil {
		return nil, fmt.Errorf(`query channel=%s chaincode=%s: %w`, channel, chaincode, err)
	}

	resProto := proto.Clone(target)

	if err = proto.Unmarshal(res.Payload, resProto); err != nil {
		return nil, fmt.Errorf(`unmarshal result to %s: %w`, reflect.TypeOf(target), err)
	}

	return resProto, nil
}
