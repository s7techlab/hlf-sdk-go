package client

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	fabPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/txwaiter"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
)

func (c *core) Invoke(
	ctx context.Context,
	channel string,
	ccName string,
	args [][]byte,
	signer msp.SigningIdentity,
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

	if signer == nil {
		signer = c.CurrentIdentity()
	}

	if endorserMSPs := tx.EndorserMSPsFromContext(ctx); len(endorserMSPs) > 0 {
		doOpts = append(doOpts, api.WithEndorsingMpsIDs(endorserMSPs))
	}

	ccAPI, err := c.Channel(channel).Chaincode(ctx, ccName)
	if err != nil {
		return nil, "", err
	}

	res, txID, err := ccAPI.Invoke(string(args[0])).
		ArgBytes(args[1:]).
		WithIdentity(signer).
		Transient(transient).
		Do(ctx, doOpts...)
	if err != nil {
		return nil, "", err
	}

	return res, txID, nil
}

func (c *core) Query(
	ctx context.Context,
	channel string,
	chaincode string,
	args [][]byte,
	identity msp.SigningIdentity,
	transient map[string][]byte,
) (*fabPeer.Response, error) {
	if identity == nil {
		identity = c.CurrentIdentity()
	}

	peer, err := c.PeerPool().FirstReadyPeer(identity.GetMSPIdentifier())
	if err != nil {
		return nil, err
	}

	return peer.Query(ctx, channel, chaincode, args, identity, transient)
}

func (c *core) Events(
	ctx context.Context,
	channel string,
	chaincode string,
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

	peer, err := c.PeerPool().FirstReadyPeer(identity.GetMSPIdentifier())
	if err != nil {
		return nil, nil, err
	}

	return peer.Events(ctx, channel, chaincode, identity, blockRange...)
}

func (c *core) Blocks(
	ctx context.Context,
	channel string,
	identity msp.SigningIdentity,
	blockRange ...int64,
) (blocks <-chan *common.Block, closer func() error, _ error) {
	if identity == nil {
		identity = c.CurrentIdentity()
	}

	peer, err := c.PeerPool().FirstReadyPeer(identity.GetMSPIdentifier())
	if err != nil {
		return nil, nil, err
	}

	return peer.Blocks(ctx, channel, identity, blockRange...)
}
