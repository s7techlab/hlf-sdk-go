package subs

import (
	"context"
	"fmt"
	//"log"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/logger"
	"github.com/s7techlab/hlf-sdk-go/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type blockSubscription struct {
	log         *zap.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	channelName string
	identity    msp.SigningIdentity
	conn        *grpc.ClientConn
	client      peer.Deliver_DeliverClient
	blockChan   chan *common.Block
	errChan     chan error
	startPos    *orderer.SeekPosition
	stopPos     *orderer.SeekPosition
}

func (b *blockSubscription) handleSubscription() {

	log := b.log.Named(`handleSubscription`)

	log.Debug(`Starting subscription`)
	defer log.Debug(`Closing subscription`)

handleLoop:
	for {
		select {
		case <-b.ctx.Done():
			log.Debug(`Caught context.Done`)
			return
		default:
			ev, err := b.client.Recv()
			log.Debug(`Got new DeliverResponse`)
			if err != nil {
				if err == context.Canceled {
					log.Debug(`Got context.Canced`)
					return
				}
				log.Error(`Subscription error`, zap.Error(err))
				b.errChan <- &api.GRPCStreamError{Err: err}
				continue handleLoop
			}

			log.Debug(`Switch DeliverResponse Type`)
			switch event := ev.Type.(type) {
			case *peer.DeliverResponse_Block:
				log.Debug(`Got DeliverResponse_Block`,
					zap.Uint64(`number`, event.Block.Header.Number),
					zap.ByteString(`hash`, event.Block.Header.DataHash),
					zap.ByteString(`prevHash`, event.Block.Header.PreviousHash),
				)
				log.Debug(`Sending block to blockChan`)
				b.blockChan <- event.Block
				log.Debug(`Sent block to blockChan`)
			default:
				log.Debug(`Got DeliverResponse UnknownType`, zap.Reflect(`type`, ev.Type))
				b.errChan <- &api.UnknownEventTypeError{Type: fmt.Sprintf("%v", ev.Type)}
				log.Debug(`Sent err to errChan`)
			}
		}
	}
}

func (b *blockSubscription) Blocks() (chan *common.Block, error) {
	log := b.log.Named(`Blocks`)

	var err error

	log.Debug(`Initializing new DeliverClient`)
	if b.client, err = peer.NewDeliverClient(b.conn).Deliver(b.ctx); err != nil {
		log.Error(`Initialization of DeliverClient failed`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to get deliver client`)
	}

	log.Debug(`Getting seekEnvelope for DeliverClient`)
	if env, err := util.SeekEnvelope(b.channelName, b.startPos, b.stopPos, b.identity); err != nil {
		log.Error(`Getting seekEnvelope failed`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to get seek envelope`)
	} else {
		log.Debug(`Got seekEnvelope`, zap.ByteString(`payload`, env.Payload), zap.ByteString(`signature`, env.Signature))
		log.Debug(`Sending seekEnvelope with DeliverClient`)
		if err = b.client.Send(env); err != nil {
			log.Error(`Sending seekEnvelope failed`, zap.Error(err))
			return nil, errors.Wrap(err, `failed to send seek envelope`)
		}
	}

	log.Debug(`Starting handleSubscription`)
	go b.handleSubscription()

	return b.blockChan, nil
}

func (b *blockSubscription) Errors() chan error {
	return b.errChan
}

func (b *blockSubscription) Close() error {

	log := b.log.Named(`Close`)

	log.Debug(`Cancelling context`)
	b.cancel()

	log.Debug(`Closing errChan`)
	close(b.errChan)

	log.Debug(`Closing blockChan`)
	close(b.blockChan)

	log.Debug(`Trying to CloseSend of DeliverClient`)
	return b.client.CloseSend()
}

func NewBlockSubscription(ctx context.Context, channelName string, identity msp.SigningIdentity, conn *grpc.ClientConn, seekOpt ...api.EventCCSeekOption) api.BlockSubscription {
	var startPos, stopPos *orderer.SeekPosition

	log := logger.DefaultLogger.
		Named(`BlockSubscription`).
		With(zap.String(`channel`, channelName))

	if len(seekOpt) > 0 {
		startPos, stopPos = seekOpt[0]()
		log.Debug(`Using presented seekOpts`, zap.Reflect(`startPos`, startPos), zap.Reflect(`stopPos`, stopPos))
	} else {
		startPos, stopPos = api.SeekNewest()()
		log.Debug(`Using default seekOpts`, zap.Reflect(`startPos`, startPos), zap.Reflect(`stopPos`, stopPos))
	}

	ctx, cancel := context.WithCancel(ctx)

	return &blockSubscription{
		log:         log,
		ctx:         ctx,
		cancel:      cancel,
		channelName: channelName,
		identity:    identity,
		conn:        conn,
		blockChan:   make(chan *common.Block),
		errChan:     make(chan error),
		startPos:    startPos,
		stopPos:     stopPos,
	}
}
