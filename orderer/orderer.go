package orderer

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hyperledger/fabric/protos/common"
	fabricOrderer "github.com/hyperledger/fabric/protos/orderer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/util"
)

type ErrUnexpectedStatus struct {
	status common.Status
}

func (e *ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status: %s", e.status.String())
}

type orderer struct {
	uri             string
	conn            *grpc.ClientConn
	ctx             context.Context
	cancel          context.CancelFunc
	connMx          sync.Mutex
	broadcastClient fabricOrderer.AtomicBroadcastClient
	grpcOptions     []grpc.DialOption
}

func (o *orderer) Broadcast(ctx context.Context, envelope *common.Envelope) (resp *fabricOrderer.BroadcastResponse, err error) {
	cli, err := o.broadcastClient.Broadcast(ctx)
	if err != nil {
		err = errors.Wrap(err, `failed to initialize broadcast client`)
		return
	}

	defer func() { err = cli.CloseSend() }()

	if err = cli.Send(envelope); err != nil {
		err = errors.Wrap(err, `failed to send envelope`)
		return
	}

	if resp, err = cli.Recv(); err != nil {
		err = errors.Wrap(err, `failed to receive response`)
		return
	} else {
		if resp.Status != common.Status_SUCCESS {
			err = &ErrUnexpectedStatus{status: resp.Status}
			return
		}
	}

	return
}

func (o *orderer) Deliver(ctx context.Context, envelope *common.Envelope) (block *common.Block, err error) {
	cli, err := o.broadcastClient.Deliver(ctx)
	if err != nil {
		err = errors.Wrap(err, `failed to initialize deliver client`)
		return
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
				err = errors.Wrap(errR, `failed to receive response`)
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
		err = errors.Wrap(err, `failed to send envelope`)
		return
	}

	err = cli.CloseSend()
	<-waitc
	return
}

func (o *orderer) initBroadcastClient() error {
	var err error
	if o.conn == nil {
		o.connMx.Lock()
		defer o.connMx.Unlock()
		if o.conn, err = grpc.DialContext(o.ctx, o.uri, o.grpcOptions...); err != nil {
			return errors.Wrap(err, `failed to initialize grpc connection`)
		}
	}

	o.broadcastClient = fabricOrderer.NewAtomicBroadcastClient(o.conn)

	return nil
}

func New(c config.ConnectionConfig, log *zap.Logger) (api.Orderer, error) {
	l := log.Named(`New`)
	opts, err := util.NewGRPCOptionsFromConfig(c, log)
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

	return NewFromGRPC(ctx, conn, opts...)
}

// NewFromGRPC allows to initialize orderer from existing GRPC connection
func NewFromGRPC(ctx context.Context, conn *grpc.ClientConn, grpcOptions ...grpc.DialOption) (api.Orderer, error) {
	obj := &orderer{
		uri:         conn.Target(),
		conn:        conn,
		ctx:         ctx,
		grpcOptions: grpcOptions,
	}

	if err := obj.initBroadcastClient(); err != nil {
		return nil, errors.Wrap(err, `failed to initialize BroadcastClient`)
	}

	return obj, nil
}
