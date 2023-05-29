package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	fabricOrderer "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protoutil"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	grpcclient "github.com/s7techlab/hlf-sdk-go/client/grpc"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
)

const (
	OrdererDefaultDialTimeout = 5 * time.Second
)

type ErrUnexpectedStatus struct {
	status  common.Status
	message string
}

func (e *ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status: %s. message: %v", e.status.String(), e.message)
}

type Orderer struct {
	uri             string
	conn            *grpc.ClientConn
	broadcastClient fabricOrderer.AtomicBroadcastClient
}

func NewOrderer(dialCtx context.Context, c config.ConnectionConfig, logger *zap.Logger) (*Orderer, error) {
	opts, err := grpcclient.OptionsFromConfig(c, logger)
	if err != nil {
		return nil, fmt.Errorf(`get orderer GRPC options: %w`, err)
	}

	// Dial shoould always has timeout
	ctxDeadline, exists := dialCtx.Deadline()
	if !exists {
		dialTimeout := c.Timeout.Duration
		if dialTimeout == 0 {
			dialTimeout = OrdererDefaultDialTimeout
		}

		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(dialCtx, dialTimeout)
		defer cancel()

		ctxDeadline, _ = dialCtx.Deadline()
	}

	logger.Debug(`dial to orderer`, zap.String(`host`, c.Host), zap.Time(`context deadline`, ctxDeadline))
	conn, err := grpc.DialContext(dialCtx, c.Host, opts.Dial...)
	if err != nil {
		return nil, fmt.Errorf(`dial to orderer=: %w`, err)
	}

	return NewOrdererFromGRPC(conn)
}

// NewOrdererFromGRPC allows initializing orderer from existing GRPC connection
func NewOrdererFromGRPC(conn *grpc.ClientConn) (*Orderer, error) {
	orderer := &Orderer{
		uri:             conn.Target(),
		conn:            conn,
		broadcastClient: fabricOrderer.NewAtomicBroadcastClient(conn),
	}

	return orderer, nil
}

func (o *Orderer) Broadcast(ctx context.Context, envelope *common.Envelope) (resp *fabricOrderer.BroadcastResponse, err error) {
	cli, err := o.broadcastClient.Broadcast(ctx)
	if err != nil {
		err = fmt.Errorf(`initialize broadcast client: %w`, err)
		return
	}

	defer func() {
		if cErr := cli.CloseSend(); cErr != nil {
			cErr = fmt.Errorf(`close client: %w`, cErr)
			if err == nil {
				err = cErr
			} else {
				err = fmt.Errorf(`%s: %w`, err.Error(), cErr)
			}
		}
	}()

	if err = cli.Send(envelope); err != nil {
		err = fmt.Errorf(`send envelope: %w`, err)
		return
	}

	if resp, err = cli.Recv(); err != nil {
		err = fmt.Errorf(`receive response: %w`, err)
		return
	} else {
		if resp.Status != common.Status_SUCCESS {
			err = &ErrUnexpectedStatus{
				status:  resp.Status,
				message: resp.Info,
			}
			return
		}
	}

	return
}

func (o *Orderer) Deliver(ctx context.Context, envelope *common.Envelope) (block *common.Block, err error) {
	cli, deliverErr := o.broadcastClient.Deliver(ctx)
	if deliverErr != nil {
		return nil, fmt.Errorf(`initialize deliver client: %w`, err)
	}

	waitc := make(chan struct{}, 0)

	go func() {
		defer close(waitc)
		for {
			resp, errR := cli.Recv()
			if errR == io.EOF {
				return
			}

			if errR != nil {
				err = fmt.Errorf(`receive response: %w`, errR)
				return
			}

			switch respType := resp.Type.(type) {
			case *fabricOrderer.DeliverResponse_Status:
				if respType.Status != common.Status_SUCCESS {
					err = &ErrUnexpectedStatus{status: respType.Status}
				} else {
					err = nil
					return
				}
			case *fabricOrderer.DeliverResponse_Block:
				block = respType.Block
				return
			}
		}
	}()

	if err = cli.Send(envelope); err != nil {
		err = fmt.Errorf(`send envelope: %w`, err)
		return
	}

	err = cli.CloseSend()
	<-waitc
	return
}

// GetConfigBlock returns config block by channel name
func (o *Orderer) GetConfigBlock(ctx context.Context, signer msp.SigningIdentity, channelName string) (*common.Block, error) {
	startPos, endPos := api.SeekNewest()()

	seekEnvelope, err := tx.NewSeekBlockEnvelope(channelName, signer, startPos, endPos, nil)
	if err != nil {
		return nil, fmt.Errorf(`create seek envelope: %w`, err)
	}

	lastBlock, err := o.Deliver(ctx, seekEnvelope)
	if err != nil {
		return nil, fmt.Errorf(`fetch last block: %w`, err)
	}

	blockId, err := protoutil.GetLastConfigIndexFromBlock(lastBlock)
	if err != nil {
		return nil, fmt.Errorf(`get last config index fron block: %w`, err)
	}

	startPos, endPos = api.SeekSingle(blockId)()

	seekEnvelope, err = tx.NewSeekBlockEnvelope(channelName, signer, startPos, endPos, nil)
	if err != nil {
		return nil, fmt.Errorf(`create seek envelope for last config block: %w`, err)
	}

	configBlock, err := o.Deliver(ctx, seekEnvelope)
	if err != nil {
		return nil, fmt.Errorf(`fetch block with config: %w`, err)
	}

	return configBlock, nil
}
