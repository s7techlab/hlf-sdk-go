package deliver

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/deliver/subs"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
)

func New(deliverCli peer.DeliverClient, identity msp.SigningIdentity) *deliverImpl {
	return &deliverImpl{
		cli:      deliverCli,
		identity: identity,
	}
}

type deliverImpl struct {
	cli      peer.DeliverClient
	identity msp.SigningIdentity
}

var (
	_ api.DeliverClient = &deliverImpl{}
)

type GetBlockerInfo interface {
	GetBlockByTxID(ctx context.Context, channelName string, txID string) (*common.Block, error)
}

type subscribeEventOption struct {
	fromTx   string
	seekOpts []api.EventCCSeekOption
	qscc     GetBlockerInfo
}

func newEventDefaultOptions() *subscribeEventOption {
	return &subscribeEventOption{
		fromTx: ``,
		seekOpts: []api.EventCCSeekOption{
			api.SeekOldest(),
		},
	}
}

func FromTxID(qscc GetBlockerInfo, txID string) func(*subscribeEventOption) error {
	return func(opt *subscribeEventOption) error {
		if len(txID) == 0 {
			return nil
		} else if qscc == nil {
			return errors.New(`GetBlockerInfo must be set for txid filter`)
		}

		opt.fromTx = txID
		opt.qscc = qscc
		return nil
	}
}

// WithDefaultSeek need if fromTxID is empty
func WithDefaultSeek(seekOpts ...api.EventCCSeekOption) func(*subscribeEventOption) error {
	return func(opt *subscribeEventOption) error {
		if len(seekOpts) > 0 {
			opt.seekOpts = seekOpts
		}
		return nil
	}
}

func WithGetBlockByTx(seekOpts ...api.EventCCSeekOption) func(*subscribeEventOption) {
	return func(opt *subscribeEventOption) {
		if len(seekOpts) > 0 {
			opt.seekOpts = seekOpts
		}
	}
}

// SubscribeEvents it is just once helper for save to api version today
func (d *deliverImpl) SubscribeEvents(ctx context.Context, channelName string, ccName string, setOpts ...func(*subscribeEventOption) error) (api.EventCCSubscription, error) {

	options := newEventDefaultOptions()

	for _, setOpt := range setOpts {
		if err := setOpt(options); err != nil {
			return nil, err
		}
	}

	events := subs.NewEventSubscription(ccName, options.fromTx)

	if len(options.fromTx) > 0 {
		b, err := options.qscc.GetBlockByTxID(ctx, channelName, options.fromTx)
		if err != nil {
			return nil, err
		}

		options.seekOpts = []api.EventCCSeekOption{
			api.SeekRange(b.Header.Number, math.MaxUint64),
		}
	}

	sub, err := d.handleSubscription(ctx, channelName, events.Handler, options.seekOpts...)
	if err != nil {
		return nil, err
	}

	return events.Serve(sub, sub.readyForHandling), nil
}

func (d *deliverImpl) SubscribeCC(ctx context.Context, channelName string, ccName string, seekOpt ...api.EventCCSeekOption) (api.EventCCSubscription, error) {
	events := subs.NewEventSubscription(ccName, ``)

	sub, err := d.handleSubscription(ctx, channelName, events.Handler, seekOpt...)
	if err != nil {
		return nil, err
	}

	return events.Serve(sub, sub.readyForHandling), nil
}

func (d *deliverImpl) SubscribeTx(ctx context.Context, channelName string, txID string, seekOpt ...api.EventCCSeekOption) (api.TxSubscription, error) {
	txSub := subs.NewTxSubscription(txID)
	sub, err := d.handleSubscription(ctx, channelName, txSub.Handler, seekOpt...)
	if err != nil {
		return nil, err
	}

	return txSub.Serve(sub, sub.readyForHandling), nil
}

func (d *deliverImpl) SubscribeBlock(ctx context.Context, channelName string, seekOpt ...api.EventCCSeekOption) (api.BlockSubscription, error) {
	blocker := subs.NewBlockSubscription()

	sub, err := d.handleSubscription(ctx, channelName, blocker.Handler, seekOpt...)
	if err != nil {
		return nil, err
	}

	return blocker.Serve(sub, sub.readyForHandling), nil
}

func (d *deliverImpl) handleSubscription(ctx context.Context, channel string, blockHandler subs.BlockHandler, seekOpt ...api.EventCCSeekOption) (*subscriptionImpl, error) {
	var startPos, stopPos *orderer.SeekPosition

	if len(seekOpt) > 0 {
		startPos, stopPos = seekOpt[0]()
	} else {
		startPos, stopPos = api.SeekNewest()()
	}

	seekEnvelope, err := tx.NewSeekBlockEnvelope(channel, d.identity, startPos, stopPos)
	if err != nil {
		return nil, fmt.Errorf(`seek envelope: %w`, err)
	}

	subCtx, stopSub := context.WithCancel(ctx)

	stream, err := d.cli.Deliver(subCtx)
	if err != nil {
		stopSub()
		return nil, errors.Wrap(err, `failed to open deliver stream`)
	}

	err = stream.Send(seekEnvelope)
	if err != nil {
		stopSub()
		return nil, errors.Wrap(err, `failed to send seek envelope to stream`)
	}

	return makeSubscription(subCtx, stopSub, stream, blockHandler), nil
}

func makeSubscription(ctx context.Context, stop context.CancelFunc, stream peer.Deliver_DeliverClient, blockHandler subs.BlockHandler) *subscriptionImpl {
	s := &subscriptionImpl{
		ctx:          ctx,
		stop:         stop,
		stream:       stream,
		blockHandler: blockHandler,
		once:         new(sync.Once),
		err:          make(chan error, 1),  // only one error
		done:         make(chan *struct{}), // done will be closed after finished sub.handle
		up:           make(chan *struct{}),
		run:          make(chan *struct{}),
	}

	go s.handle()
	<-s.up

	return s
}

type subscriptionImpl struct {
	ctx          context.Context
	stop         context.CancelFunc
	blockHandler subs.BlockHandler
	stream       peer.Deliver_DeliverClient
	err          chan error
	once         *sync.Once
	done         chan *struct{}
	up           chan *struct{}
	run          chan *struct{}
}

func (s *subscriptionImpl) handle() {
	defer func() { _ = s.Close() }()
	defer close(s.done)
	close(s.up)
	// wait of set to handler
	<-s.run

	ctx := s.stream.Context()
	for {
		ev, err := s.stream.Recv()
		if err == io.EOF {
			s.blockHandler(nil)
			return
		}

		if err != nil {
			s.err <- err
			s.blockHandler(nil) // if arg is nil, events channel will be closed
			return
		}

		switch event := ev.Type.(type) {
		case *peer.DeliverResponse_Block:
			select {
			case <-ctx.Done():
				s.err <- ctx.Err()
				return
			default:
				if skip := s.blockHandler(event.Block); skip {
					return
				}
			}
		default:
			continue
		}
	}
}

func (s *subscriptionImpl) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *subscriptionImpl) readyForHandling() {
	close(s.run)
}

func (s *subscriptionImpl) Err() <-chan error {
	return s.err
}

func (s *subscriptionImpl) Errors() chan error {
	return s.err
}

func (s *subscriptionImpl) Close() error {
	var err error

	s.once.Do(func() {
		err = s.stream.CloseSend()
		s.stop()
		//wait of stop handler
		<-s.done
		// close all channels
		close(s.err)
	})

	return err
}
