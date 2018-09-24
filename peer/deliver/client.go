package deliver

import (
	"context"
	"sync"

	"github.com/s7techlab/hlf-sdk-go/peer"

	"go.uber.org/zap"

	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver/subs"
	"google.golang.org/grpc"
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

type deliverClient struct {
	log      *zap.Logger
	uri      string
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

func (e *deliverClient) Close() error {
	return e.conn.Close()
}

func NewDeliverClient(c config.PeerConfig, identity msp.SigningIdentity, log *zap.Logger) (api.DeliverClient, error) {
	l := log.Named(`DeliverClient`)

	conn, err := peer.NewGRPCFromConfig(c, l)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize GRPC connection`)
	}
	return NewFromGRPC(conn, identity, l), nil
}

// NewFromGRPC allows to initialize orderer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, identity msp.SigningIdentity, log *zap.Logger) api.DeliverClient {
	l := log.Named(`NewFromGRPC`)
	l.Debug(`Using presented GRPC connection`, zap.String(`target`, conn.Target()))
	l.Debug(`Using presented identity`, zap.String(`msp_id`, identity.GetMSPIdentifier()), zap.String(`id`, identity.GetIdentifier().Id))
	return &deliverClient{
		log:      l,
		uri:      conn.Target(),
		conn:     conn,
		identity: identity,
	}
}
