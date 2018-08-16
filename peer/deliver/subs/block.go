package subs

import (
	"context"
	"log"

	"fmt"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/util"
	"google.golang.org/grpc"
)

type blockSubscription struct {
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
	closeChan   chan struct{}
}

func (b *blockSubscription) handleSubscription() {

handleLoop:
	for {
		select {
		case <-b.closeChan:
			return
		default:
			ev, err := b.client.Recv()
			if err != nil {
				if err == context.Canceled {
					return
				}
				b.errChan <- &api.GRPCStreamError{Err: err}
				continue handleLoop
			}

			switch event := ev.Type.(type) {
			case *peer.DeliverResponse_Block:
				b.blockChan <- event.Block
			case *peer.DeliverResponse_FilteredBlock:
				b.errChan <- &api.UnknownEventTypeError{Type: fmt.Sprintf("%v", ev.Type)}
			}
		}
	}
}

func (b *blockSubscription) Blocks() (chan *common.Block, error) {
	var err error
	if b.client, err = peer.NewDeliverClient(b.conn).Deliver(b.ctx); err != nil {
		return nil, errors.Wrap(err, `failed to get deliver client`)
	}

	if env, err := util.SeekEnvelope(b.channelName, b.startPos, b.stopPos, b.identity); err != nil {
		return nil, errors.Wrap(err, `failed to get seek envelope`)
	} else {
		if err = b.client.Send(env); err != nil {
			return nil, errors.Wrap(err, `failed to send seek envelope`)
		}
	}

	go b.handleSubscription()

	return b.blockChan, nil
}

func (b *blockSubscription) Errors() chan error {
	return b.errChan
}

func (b *blockSubscription) Close() error {
	log.Println(`Cancelling context`)
	b.cancel()

	log.Println(`Closing handleSubscription`)
	b.closeChan <- struct{}{}

	log.Println(`Closing errChan`)
	close(b.errChan)

	log.Println(`Closing blockChan`)
	close(b.blockChan)

	return b.client.CloseSend()
}

func NewBlockSubscription(channelName string, identity msp.SigningIdentity, conn *grpc.ClientConn, seekOpt ...api.EventCCSeekOption) api.BlockSubscription {
	var startPos, stopPos *orderer.SeekPosition

	if len(seekOpt) > 0 {
		startPos, stopPos = seekOpt[0]()
	} else {
		startPos, stopPos = api.SeekNewest()()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &blockSubscription{
		ctx:         ctx,
		cancel:      cancel,
		channelName: channelName,
		identity:    identity,
		conn:        conn,
		blockChan:   make(chan *common.Block),
		errChan:     make(chan error),
		startPos:    startPos,
		stopPos:     stopPos,
		closeChan:   make(chan struct{}),
	}
}
