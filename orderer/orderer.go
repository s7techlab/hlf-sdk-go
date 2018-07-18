package orderer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric/protos/common"
	fabricOrderer "github.com/hyperledger/fabric/protos/orderer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

var (
	errTimeoutExceeded = errors.New(`timeout exceeded`)
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
	connMx          sync.Mutex
	timeout         time.Duration
	broadcastClient fabricOrderer.AtomicBroadcastClient
	grpcOptions     []grpc.DialOption
}

func (o *orderer) Broadcast(envelope *common.Envelope) (*fabricOrderer.BroadcastResponse, error) {
	cli, err := o.broadcastClient.Broadcast(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize broadcast client`)
	}
	defer cli.CloseSend()

	if err = cli.Send(envelope); err != nil {
		return nil, errors.Wrap(err, `failed to send envelope`)
	}

	if resp, err := cli.Recv(); err != nil {
		return nil, errors.Wrap(err, `failed to receive response`)
	} else {
		if resp.Status != common.Status_SUCCESS {
			return nil, &ErrUnexpectedStatus{status: resp.Status}
		}
		return resp, nil
	}
}

func (o *orderer) Deliver(envelope *common.Envelope) (*common.Block, error) {
	cli, err := o.broadcastClient.Deliver(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize deliver client`)
	}
	defer cli.CloseSend()

	if err = cli.Send(envelope); err != nil {
		return nil, errors.Wrap(err, `failed to send envelope`)
	}

	var block *common.Block

	timer := time.NewTimer(o.timeout)

	for {
		select {
		case <-timer.C:
			return nil, errTimeoutExceeded
		default:
			if resp, err := cli.Recv(); err != nil {
				return nil, errors.Wrap(err, `failed to receive response`)
			} else {
				switch respType := resp.Type.(type) {
				case *fabricOrderer.DeliverResponse_Status:
					if respType.Status != common.Status_SUCCESS {
						return nil, &ErrUnexpectedStatus{status: respType.Status}
					} else {
						return block, nil
					}
				case *fabricOrderer.DeliverResponse_Block:
					block = respType.Block

				}
			}
		}
	}
}

func (o *orderer) initBroadcastClient() error {
	var err error
	if o.conn == nil {
		o.connMx.Lock()
		defer o.connMx.Unlock()
		if o.conn, err = grpc.Dial(o.uri, o.grpcOptions...); err != nil {
			return errors.Wrap(err, `failed to initialize grpc connection`)
		}
		o.broadcastClient = fabricOrderer.NewAtomicBroadcastClient(o.conn)
	}
	return nil
}

func New(c config.OrdererConfig) (api.Orderer, error) {
	var err error
	o := orderer{uri: c.Host, grpcOptions: make([]grpc.DialOption, 0)}
	if c.Tls.Enabled {
		if ts, err := credentials.NewClientTLSFromFile(c.Tls.CertPath, ``); err != nil {
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			o.grpcOptions = append(o.grpcOptions, grpc.WithTransportCredentials(ts))
		}
	} else {
		o.grpcOptions = append(o.grpcOptions, grpc.WithInsecure())
	}

	if c.GRPC.KeepAlive != nil {
		o.grpcOptions = append(o.grpcOptions, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Duration(c.GRPC.KeepAlive.Time) * time.Second,
			Timeout: time.Duration(c.GRPC.KeepAlive.Timeout) * time.Second,
		}))
	}

	o.grpcOptions = append(o.grpcOptions, grpc.WithBlock(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	if o.timeout, err = time.ParseDuration(c.Timeout); err != nil {
		return nil, errors.Wrap(err, `failed to parse timeout duration`)
	}

	if err = o.initBroadcastClient(); err != nil {
		return nil, errors.Wrap(err, `failed to initialize BroadcastClient`)
	}

	return &o, nil
}
