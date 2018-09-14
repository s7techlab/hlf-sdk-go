package deliver

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"time"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver/subs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

type deliverClient struct {
	log      *zap.Logger
	uri      string
	opts     []grpc.DialOption
	identity msp.SigningIdentity
	conn     *grpc.ClientConn
	connMx   sync.Mutex
}

func (e *deliverClient) SubscribeCC(ctx context.Context, channelName string, ccName string, seekOpt ...api.EventCCSeekOption) api.EventCCSubscription {
	return subs.NewEventSubscription(ctx, channelName, ccName, e.identity, e.conn, e.log, seekOpt...)
}

func (e *deliverClient) SubscribeTx(ctx context.Context, channelName string, txId api.ChaincodeTx) api.TxSubscription {
	return subs.NewTxSubscription(ctx, txId, channelName, e.identity, e.conn, e.log, api.SeekNewest())
}

func (e *deliverClient) SubscribeBlock(ctx context.Context, channelName string, seekOpt ...api.EventCCSeekOption) api.BlockSubscription {
	return subs.NewBlockSubscription(ctx, channelName, e.identity, e.conn, e.log, seekOpt...)
}

func (e *deliverClient) initConnection() error {
	var err error

	e.connMx.Lock()
	defer e.connMx.Unlock()

	if e.conn == nil {
		if e.conn, err = grpc.Dial(e.uri, e.opts...); err != nil {
			return errors.Wrap(err, `failed to initialize grpc connection`)
		}
	}
	return nil
}

func (e *deliverClient) Close() error {
	return e.conn.Close()
}

func NewDeliverClient(config config.PeerConfig, identity msp.SigningIdentity, log *zap.Logger, grpcOptions ...grpc.DialOption) (api.DeliverClient, error) {
	l := log.Named(`DeliverClient`)

	var err error
	cli := &deliverClient{
		log:      l,
		uri:      config.Host,
		opts:     grpcOptions,
		identity: identity,
	}

	if config.Tls.Enabled {
		l.Debug(`Using TLS connection`)
		if ts, err := credentials.NewClientTLSFromFile(config.Tls.CertPath, ``); err != nil {
			l.Debug(`Failed to initiate TLS credentials`, zap.Error(err))
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			cli.opts = append(cli.opts, grpc.WithTransportCredentials(ts))
		}
	} else {
		l.Debug(`Using insecure connection`)
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
		l.Debug(`Failed to init DeliverClient`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to initialize DeliverClient`)
	}

	return cli, nil
}

// NewFromGRPC allows to initialize orderer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, identity msp.SigningIdentity, log *zap.Logger, grpcOptions ...grpc.DialOption) api.DeliverClient {
	l := log.Named(`DeliverClient`)
	l.Debug(`Using presented GRPC connection`, zap.String(`target`, conn.Target()))
	l.Debug(`Using presented identity`, zap.String(`msp_id`, identity.GetMSPIdentifier()), zap.String(`id`, identity.GetIdentifier().Id))
	return &deliverClient{
		log:      l,
		uri:      conn.Target(),
		conn:     conn,
		opts:     grpcOptions,
		identity: identity,
	}
}
