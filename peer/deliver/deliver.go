package deliver

import (
	"context"
	"io"
	"sync"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver/subs"
	"github.com/s7techlab/hlf-sdk-go/util"
)

// New
func New(delivercli peer.DeliverClient, identity msp.SigningIdentity) api.DeliverClient {
	return &deliverImpl{
		cli:      delivercli,
		identity: identity,
	}
}

type deliverImpl struct {
	cli      peer.DeliverClient
	identity msp.SigningIdentity
}

func (d *deliverImpl) SubscribeCC(ctx context.Context, channelName string, ccName string, seekOpt ...api.EventCCSeekOption) (api.EventCCSubscription, error) {
	events := subs.NewEventSubscription(ccName)

	sub, err := d.handleSubscription(ctx, channelName, events.Handler)
	if err != nil {
		return nil, err
	}

	return events.Serve(sub), nil
}

func (d *deliverImpl) SubscribeTx(ctx context.Context, channelName string, txId api.ChaincodeTx, seekOpt ...api.EventCCSeekOption) (api.TxSubscription, error) {
	txSub := subs.NewTxSubscription(txId)
	sub, err := d.handleSubscription(ctx, channelName, txSub.Handler, seekOpt...)
	if err != nil {
		return nil, err
	}

	return txSub.Serve(sub), nil
}

func (d *deliverImpl) SubscribeBlock(ctx context.Context, channelName string, seekOpt ...api.EventCCSeekOption) (api.BlockSubscription, error) {
	blocker := subs.NewBlockSubscription()

	sub, err := d.handleSubscription(ctx, channelName, blocker.Handler, seekOpt...)
	if err != nil {
		return nil, err
	}

	return blocker.Serve(sub), nil
}

func (d *deliverImpl) handleSubscription(ctx context.Context, channel string, blockHandler subs.BlockHandler, seekOpt ...api.EventCCSeekOption) (*subscriptionImpl, error) {
	var startPos, stopPos *orderer.SeekPosition

	if len(seekOpt) > 0 {
		startPos, stopPos = seekOpt[0]()
	} else {
		startPos, stopPos = api.SeekNewest()()
	}

	seek, err := util.SeekEnvelope(channel, startPos, stopPos, d.identity)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get seek envelope`)
	}

	stream, err := d.cli.Deliver(ctx)
	if err != nil {
		return nil, errors.Wrap(err, `failed to open deliver stream`)
	}

	err = stream.Send(seek)
	if err != nil {
		return nil, errors.Wrap(err, `failed to send seek envelope to stream`)
	}

	return makeSubscription(stream, blockHandler), nil
}

func makeSubscription(stream peer.Deliver_DeliverClient, blockHandler subs.BlockHandler) *subscriptionImpl {
	s := &subscriptionImpl{
		stream:       stream,
		blockHandler: blockHandler,
		once:         new(sync.Once),
		err:          make(chan error, 1),  // only one error
		done:         make(chan *struct{}), // done will be closed after finished sub.handle
		up:           make(chan *struct{}),
	}

	go s.handle()
	<-s.up

	return s
}

type subscriptionImpl struct {
	blockHandler subs.BlockHandler
	stream       peer.Deliver_DeliverClient
	err          chan error
	once         *sync.Once
	done         chan *struct{}
	up           chan *struct{}
}

func (s *subscriptionImpl) handle() {
	defer s.Close()
	defer close(s.done)
	close(s.up)

	ctx := s.stream.Context()
	for {
		ev, err := s.stream.Recv()
		if err == io.EOF {
			s.blockHandler(nil)
			return
		}

		if err != nil {
			s.err <- err
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
					continue
				}
			}
		default:
			continue
		}
	}
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
		//wait of stop handler
		<-s.done
		// close all channels
		close(s.err)
	})

	return err
}
