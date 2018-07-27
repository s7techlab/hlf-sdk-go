package deliver

import (
	"sync"

	"time"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver/subs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

type deliverClient struct {
	uri      string
	opts     []grpc.DialOption
	identity msp.SigningIdentity
	conn     *grpc.ClientConn
	connMx   sync.Mutex
}

func (e *deliverClient) SubscribeCC(channelName string, ccName string, seekOpt ...api.EventCCSeekOption) api.EventCCSubscription {
	return subs.NewEventSubscription(channelName, ccName, e.identity, e.conn, seekOpt...)
}

func (e *deliverClient) SubscribeTx(channelName string, txId api.ChaincodeTx) api.TxSubscription {
	return subs.NewTxSubscription(txId, channelName, e.identity, e.conn, api.SeekNewest())
}

func (e *deliverClient) SubscribeBlock(channelName string, seekOpt ...api.EventCCSeekOption) api.BlockSubscription {
	return subs.NewBlockSubscription(channelName, e.identity, e.conn, seekOpt...)
}

func (e *deliverClient) initConnection() error {
	var err error
	e.connMx.Lock()
	defer e.connMx.Unlock()
	if e.conn == nil || e.conn.GetState() == connectivity.Shutdown {
		if e.conn, err = grpc.Dial(e.uri, e.opts...); err != nil {
			return errors.Wrap(err, `failed to initialize grpc connection`)
		}
	}
	return nil
}

func (e *deliverClient) Close() error {
	return e.conn.Close()
}

func NewDeliverClient(config config.PeerConfig, identity msp.SigningIdentity, grpcOptions ...grpc.DialOption) (api.DeliverClient, error) {
	var err error
	cli := &deliverClient{
		uri:      config.Host,
		opts:     grpcOptions,
		identity: identity,
	}

	if config.Tls.Enabled {
		if ts, err := credentials.NewClientTLSFromFile(config.Tls.CertPath, ``); err != nil {
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			cli.opts = append(cli.opts, grpc.WithTransportCredentials(ts))
		}
	} else {
		cli.opts = append(cli.opts, grpc.WithInsecure())
	}

	// Set KeepAlive parameters if presented
	if config.GRPC.KeepAlive != nil {
		cli.opts = append(cli.opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Duration(config.GRPC.KeepAlive.Time) * time.Second,
			Timeout: time.Duration(config.GRPC.KeepAlive.Timeout) * time.Second,
		}))
	}

	cli.opts = append(cli.opts, grpc.WithBlock(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	if err = cli.initConnection(); err != nil {
		return nil, errors.Wrap(err, `failed to initialize DeliverClient`)
	}

	return cli, nil
}
