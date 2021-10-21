package orderer

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	fabricOrderer "github.com/hyperledger/fabric-protos-go/orderer"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/api/config"
	"github.com/s7techlab/hlf-sdk-go/v2/util"
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
			err = &ErrUnexpectedStatus{status: resp.Status}
			return
		}
	}

	return
}

func (o *orderer) Deliver(ctx context.Context, envelope *common.Envelope) (block *common.Block, err error) {
	cli, err := o.broadcastClient.Deliver(ctx)
	if err != nil {
		err = fmt.Errorf(`initialize deliver client: %w`, err)
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

func (o *orderer) initBroadcastClient() error {
	var err error
	if o.conn == nil {
		o.connMx.Lock()
		defer o.connMx.Unlock()
		if o.conn, err = grpc.DialContext(o.ctx, o.uri, o.grpcOptions...); err != nil {
			return fmt.Errorf(`initialize grpc connection: %w`, err)
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
		return nil, fmt.Errorf(`get GRPC options: %w`, err)
	}

	ctx, _ := context.WithTimeout(context.Background(), c.Timeout.Duration)
	conn, err := grpc.DialContext(ctx, c.Host, opts...)
	if err != nil {
		l.Error(`Failed to initialize GRPC connection`, zap.Error(err))
		return nil, fmt.Errorf(`initialize GRPC connection: %w`, err)
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
		return nil, fmt.Errorf(`initialize BroadcastClient: %w`, err)
	}

	return obj, nil
}
