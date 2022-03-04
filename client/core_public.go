package client

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	fabPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/txwaiter"
)

func (c *core) Invoke(
	ctx context.Context,
	chanName string,
	ccName string,
	args [][]byte,
	identity msp.SigningIdentity,
	transient map[string][]byte,
	txWaiterType string,
) (*fabPeer.Response, string, error) {
	var doOpts []api.DoOption

	switch txWaiterType {
	case "":
		doOpts = append(doOpts, chaincode.WithTxWaiter(txwaiter.Self))
	case api.TxWaiterSelfType:
		doOpts = append(doOpts, chaincode.WithTxWaiter(txwaiter.Self))
	case api.TxWaiterAllType:
		doOpts = append(doOpts, chaincode.WithTxWaiter(txwaiter.All))
	default:
		return nil, "", fmt.Errorf("invalid tx waiter type. got %v, available: '%v', '%v'", txWaiterType, api.TxWaiterSelfType, api.TxWaiterAllType)
	}

	if identity == nil {
		identity = c.CurrentIdentity()
	}

	ccAPI, err := c.Channel(chanName).Chaincode(ctx, ccName)
	if err != nil {
		return nil, "", err
	}

	res, tx, err := ccAPI.Invoke(string(args[0])).
		ArgBytes(args[1:]).
		WithIdentity(identity).
		Transient(transient).
		Do(ctx, doOpts...)
	if err != nil {
		return nil, "", err
	}

	return res, string(tx), nil
}

func (c *core) Query(
	ctx context.Context,
	chanName string,
	ccName string,
	args [][]byte,
	identity msp.SigningIdentity,
	transient map[string][]byte,
) (*fabPeer.Response, error) {
	if identity == nil {
		identity = c.CurrentIdentity()
	}

	ccAPI, err := c.Channel(chanName).Chaincode(ctx, ccName)
	if err != nil {
		return nil, err
	}

	argsStrings := make([]string, 0)
	for _, arg := range args {
		argsStrings = append(argsStrings, string(arg))
	}

	return ccAPI.Query(argsStrings[0], argsStrings[1:]...).
		WithIdentity(identity).
		Transient(transient).
		Do(ctx)
}

func (c *core) Events(
	ctx context.Context,
	chanName string,
	ccName string,
	identity msp.SigningIdentity,
	blockRange ...int64,
) (events chan interface {
	Event() *fabPeer.ChaincodeEvent
	Block() uint64
	TxTimestamp() *timestamp.Timestamp
}, closer func() error, err error) {
	if identity == nil {
		identity = c.CurrentIdentity()
	}

	c.logger.Debug(`[Events] block range`, zap.String("chanName", chanName), zap.Reflect(`slice`, blockRange))
	mspID := identity.GetMSPIdentifier()

	dc, err := c.PeerPool().DeliverClient(mspID, identity)
	if err != nil {
		return nil, nil, err
	}

	var seekOpts []api.EventCCSeekOption
	seekOpt, err := NewSeekOptConverter(c).ByBlockRange(ctx, chanName, blockRange...)
	if err != nil {
		return nil, nil, err
	}

	if seekOpt != nil {
		seekOpts = append(seekOpts, seekOpt)
	}

	sub, err := dc.SubscribeCC(ctx, chanName, ccName, seekOpts...)
	if err != nil {
		return nil, nil, err
	}

	return sub.EventsExtended(), sub.Close, nil
}

func (c *core) Blocks(
	ctx context.Context,
	chanName string,
	identity msp.SigningIdentity,
	blockRange ...int64,
) (blockChan <-chan *common.Block, closer func() error, err error) {
	if identity == nil {
		identity = c.CurrentIdentity()
	}

	c.logger.Debug(`[Blocks] block range`, zap.String("chanName", chanName), zap.Reflect(`slice`, blockRange))
	mspID := identity.GetMSPIdentifier()

	dc, err := c.PeerPool().DeliverClient(mspID, identity)
	if err != nil {
		return nil, nil, err
	}

	var seekOpts []api.EventCCSeekOption
	seekOpt, err := NewSeekOptConverter(c).ByBlockRange(ctx, chanName, blockRange...)
	if err != nil {
		return nil, nil, err
	}

	if seekOpt != nil {
		seekOpts = append(seekOpts, seekOpt)
	}

	bs, err := dc.SubscribeBlock(ctx, chanName, seekOpts...)
	if err != nil {
		return nil, nil, err
	}

	return bs.Blocks(), bs.Close, nil
}
