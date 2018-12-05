package deliver

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver/subs"
	"github.com/s7techlab/hlf-sdk-go/util"
)

const (
	subIndexLength = 5
)

type deliverClient struct {
	ctx      context.Context
	cancel   context.CancelFunc
	log      *zap.Logger
	uri      string
	identity msp.SigningIdentity
	conn     *grpc.ClientConn

	blockSubStore   map[string]*dcBlockSub
	blockSubStoreMx sync.Mutex
}

type dcBlockSub struct {
	ctx         context.Context
	log         *zap.Logger
	blockSub    api.BlockSubscription
	listeners   map[string]*dcBlockSubListener
	listenersMx sync.Mutex
}

type dcBlockSubListener struct {
	blockChan chan *common.Block
	errChan   chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

func (sub *dcBlockSub) addSub(ctx context.Context) (chan *common.Block, chan error, context.CancelFunc, error) {
	sub.listenersMx.Lock()
	defer sub.listenersMx.Unlock()
	log := sub.log.Named(`addSub`)
	subHash := util.RandStringBytesMaskImprSrc(subIndexLength)
	log.Debug(`Adding new sub`, zap.String(`id`, subHash))
	if _, ok := sub.listeners[subHash]; ok {
		return nil, nil, nil, errors.New(`subs hash collision`)
	} else {
		newCtx, cancel := context.WithCancel(ctx)
		newSub := dcBlockSubListener{blockChan: make(chan *common.Block), ctx: newCtx, cancel: cancel, errChan: make(chan error)}
		sub.listeners[subHash] = &newSub
		return newSub.blockChan, newSub.errChan, cancel, nil
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
						listener.errChan <- err
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
			log.Debug(`Got blockSub error`, zap.Error(err))
			if !ok {
				log.Error(`blockSub is closed`)
				return
			}
			sub.listenersMx.Lock()

			log.Debug(`Iterating over listeners`, zap.Int(`listener_count`, len(sub.listeners)))
			for _, listener := range sub.listeners {
				// TODO think about listener errChan
				if listener.ctx.Err() == nil {
					log.Debug(`Listener isn't canceled, `)
					listener.errChan <- err
				} else {
					listener.cancel()
				}
			}
			sub.listenersMx.Unlock()
		case <-sub.ctx.Done():
			return
		}
	}
}

func (e *deliverClient) SubscribeCC(ctx context.Context, channelName string, ccName string) (api.EventCCSubscription, error) {
	blockChan, errChan, stop, err := e.initializeBlockSub(e.ctx, channelName)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize block channel`)
	}
	return subs.NewEventSubscription(ctx, blockChan, errChan, stop, e.log), nil
}

func (e *deliverClient) SubscribeTx(ctx context.Context, channelName string, txId api.ChaincodeTx) (api.TxSubscription, error) {
	blockChan, errChan, stop, err := e.initializeBlockSub(e.ctx, channelName)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize block channel`)
	}
	return subs.NewTxSubscription(ctx, txId, blockChan, errChan, stop, e.log), nil
}

func (e *deliverClient) SubscribeBlock(ctx context.Context, channelName string, seekOpt ...api.EventCCSeekOption) (api.BlockSubscription, error) {
	return subs.NewBlockSubscription(ctx, channelName, e.identity, e.conn, e.log, seekOpt...)
}

func (e *deliverClient) Close() error {
	log := e.log.Named(`Close`)
	log.Debug(`Canceling context`)
	e.cancel()
	return e.conn.Close()
}

func (e *deliverClient) initializeBlockSub(ctx context.Context, channelName string) (chan *common.Block, chan error, context.CancelFunc, error) {
	log := e.log.Named(`initializeBlockSub`).With(zap.String(`channel`, channelName))
	var err error
	e.blockSubStoreMx.Lock()
	defer e.blockSubStoreMx.Unlock()
	log.Debug(`Searching blockSub`)
	if sub, ok := e.blockSubStore[channelName]; !ok {
		log.Debug(`blockSub is not found, constructing new`)
		sub = &dcBlockSub{ctx: ctx, listeners: make(map[string]*dcBlockSubListener), log: e.log.Named(`dcBlockSub`)}
		if sub.blockSub, err = e.SubscribeBlock(ctx, channelName); err != nil {
			log.Debug(`Failed to initiate blockSub`, zap.Error(err))
			return nil, nil, nil, err
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

// NewFromGRPC allows to initialize orderer from existing GRPC connection
func NewFromGRPC(ctx context.Context, conn *grpc.ClientConn, identity msp.SigningIdentity, log *zap.Logger) api.DeliverClient {
	l := log.Named(`NewFromGRPC`)
	l.Debug(`Using presented GRPC connection`, zap.String(`target`, conn.Target()))
	l.Debug(`Using presented identity`, zap.String(`msp_id`, identity.GetMSPIdentifier()), zap.String(`id`, identity.GetIdentifier().Id))

	newCtx, cancel := context.WithCancel(ctx)

	return &deliverClient{
		ctx:           newCtx,
		cancel:        cancel,
		log:           l,
		uri:           conn.Target(),
		conn:          conn,
		identity:      identity,
		blockSubStore: make(map[string]*dcBlockSub),
	}
}
