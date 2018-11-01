package deliver

import (
	"context"
	"sync"

	"github.com/thanhpk/randstr"

	"github.com/hyperledger/fabric/protos/common"

	"github.com/s7techlab/hlf-sdk-go/util"

	"go.uber.org/zap"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver/subs"
	"google.golang.org/grpc"
)

type deliverClient struct {
	log      *zap.Logger
	uri      string
	identity msp.SigningIdentity
	conn     *grpc.ClientConn

	blockSubStore   map[string]*dcBlockSub
	blockSubStoreMx sync.Mutex
}

type dcBlockSub struct {
	log         *zap.Logger
	blockSub    api.BlockSubscription
	listeners   map[string]*dcBlockSubListener
	listenersMx sync.Mutex
}

type dcBlockSubListener struct {
	blockChan chan *common.Block
	ctx       context.Context
}

func (sub *dcBlockSub) addSub(ctx context.Context) (chan *common.Block, error) {
	sub.listenersMx.Lock()
	defer sub.listenersMx.Unlock()
	log := sub.log.Named(`addSub`)
	subHash := randstr.Hex(5)
	log.Debug(`Adding new sub`, zap.String(`id`, subHash))
	if _, ok := sub.listeners[subHash]; ok {
		return nil, errors.New(`subs hash collision`)
	} else {
		newSub := dcBlockSubListener{blockChan: make(chan *common.Block), ctx: ctx}
		sub.listeners[subHash] = &newSub
		return newSub.blockChan, nil
	}
}

func (sub *dcBlockSub) handle() {
	var err error
	log := sub.log.Named(`handle`)
	defer sub.blockSub.Close()
	for {
		select {
		case block, ok := <-sub.blockSub.Blocks():
			if !ok {
				log.Error(`blockSub is closed`)
				return
			} else {
				sub.listenersMx.Lock()
				log.Debug(`Iterating over listeners`, zap.Int(`listener_count`, len(sub.listeners)))
				for key, listener := range sub.listeners {
					if err = listener.ctx.Err(); err != nil {
						log.Debug(`Listener is done`, zap.Error(err), zap.String(`subKey`, key))
						delete(sub.listeners, key)
					} else {
						log.Debug(`Sending block to sub`, zap.String(`subKey`, key))
						listener.blockChan <- block
						log.Debug(`Sent block to sub`, zap.String(`subKey`, key))
					}
				}
				sub.listenersMx.Unlock()
			}
		case err, ok := <-sub.blockSub.Errors():
			if !ok {
				return
			}
			switch err.(type) {
			case api.GRPCStreamError:
				return
			}
		}
	}
}

func (e *deliverClient) SubscribeCC(ctx context.Context, channelName string, ccName string) (api.EventCCSubscription, error) {
	blockChan, err := e.initializeBlockSub(ctx, channelName)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize block channel`)
	}
	return subs.NewEventSubscription(ctx, blockChan, e.log), nil
}

func (e *deliverClient) SubscribeTx(ctx context.Context, channelName string, txId api.ChaincodeTx) (api.TxSubscription, error) {
	blockChan, err := e.initializeBlockSub(ctx, channelName)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize block channel`)
	}
	return subs.NewTxSubscription(ctx, txId, blockChan, e.log), nil
}

func (e *deliverClient) SubscribeBlock(ctx context.Context, channelName string, seekOpt ...api.EventCCSeekOption) (api.BlockSubscription, error) {
	return subs.NewBlockSubscription(ctx, channelName, e.identity, e.conn, e.log, seekOpt...)
}

func (e *deliverClient) Close() error {
	return e.conn.Close()
}

func (e *deliverClient) initializeBlockSub(ctx context.Context, channelName string) (chan *common.Block, error) {
	log := e.log.Named(`initializeBlockSub`).With(zap.String(`channel`, channelName))
	var err error
	e.blockSubStoreMx.Lock()
	defer e.blockSubStoreMx.Unlock()
	log.Debug(`Searching blockSub`)
	if sub, ok := e.blockSubStore[channelName]; !ok {
		log.Debug(`blockSub is not found, constructing new`)
		sub = &dcBlockSub{listeners: make(map[string]*dcBlockSubListener), log: e.log.Named(`dcBlockSub`)}
		if sub.blockSub, err = e.SubscribeBlock(ctx, channelName); err != nil {
			log.Debug(`Failed to initiate blockSub`, zap.Error(err))
			return nil, err
		}

		log.Debug(`Starting to handle dcBlockSub`)
		go sub.handle()
		e.blockSubStore[channelName] = sub
		return sub.addSub(ctx)
	} else {
		log.Debug(`blockSub is found, adding to exist sub`)
		return sub.addSub(ctx)
	}
}

func NewDeliverClient(c config.ConnectionConfig, identity msp.SigningIdentity, log *zap.Logger) (api.DeliverClient, error) {
	l := log.Named(`DeliverClient`)

	opts, err := util.NewGRPCOptionsFromConfig(c, l)
	if err != nil {
		l.Error(`Failed to get GRPC options`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to get GRPC options`)
	}

	ctx, _ := context.WithTimeout(context.Background(), c.Timeout.Duration)

	conn, err := grpc.DialContext(ctx, c.Host, opts...)
	if err != nil {
		l.Error(`Failed to initialize GRPC connection`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to initialize GRPC connection`)
	}

	return NewFromGRPC(conn, identity, l), nil
}

// NewFromGRPC allows to initialize orderer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, identity msp.SigningIdentity, log *zap.Logger) api.DeliverClient {
	l := log.Named(`NewFromGRPC`)
	l.Debug(`Using presented GRPC connection`, zap.String(`target`, conn.Target()))
	l.Debug(`Using presented identity`, zap.String(`msp_id`, identity.GetMSPIdentifier()), zap.String(`id`, identity.GetIdentifier().Id))
	return &deliverClient{
		log:           l,
		uri:           conn.Target(),
		conn:          conn,
		identity:      identity,
		blockSubStore: make(map[string]*dcBlockSub),
	}
}
