package peer

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/api/config"
	"github.com/s7techlab/hlf-sdk-go/v2/peer/deliver"
	"github.com/s7techlab/hlf-sdk-go/v2/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type peer struct {
	log     *zap.Logger
	conn    *grpc.ClientConn
	timeout time.Duration
	client  fabricPeer.EndorserClient
}

var (
	defaultTimeout = 5 * time.Second
)

func (p *peer) Endorse(ctx context.Context, proposal *fabricPeer.SignedProposal, opts ...api.PeerEndorseOpt) (*fabricPeer.ProposalResponse, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	resp, err := p.client.ProcessProposal(ctx, proposal)
	if err != nil {
		return nil, err
	}

	if resp.Response.Status != shim.OK {
		return nil, api.PeerEndorseError{Status: resp.Response.Status, Message: resp.Response.Message}
	}

	return resp, nil
}

func (p *peer) DeliverClient(identity msp.SigningIdentity) (api.DeliverClient, error) {
	return deliver.New(fabricPeer.NewDeliverClient(p.conn), identity), nil
}

func (p *peer) Conn() *grpc.ClientConn {
	return p.conn
}

func (p *peer) Uri() string {
	return p.conn.Target()
}

func (p *peer) Close() error {
	return p.conn.Close()
}

func (p *peer) initEndorserClient() error {
	if p.conn == nil {
		return errors.New(`empty connection`)
	}

	if p.client == nil {
		p.client = fabricPeer.NewEndorserClient(p.conn)
	}

	return nil
}

// New returns new peer instance based on peer config
func New(c config.ConnectionConfig, log *zap.Logger) (api.Peer, error) {
	opts, err := util.NewGRPCOptionsFromConfig(c, log)
	if err != nil {
		return nil, fmt.Errorf(`grpc options from config: %w`, err)
	}

	timeout := c.Timeout.Duration
	if timeout == 0 {
		timeout = defaultTimeout
	}

	//ctx, _ := context.WithTimeout(context.Background(), c.Timeout.Duration)
	log.Debug(`dial to peer`, zap.String(`host`, c.Host), zap.Duration(`timeout`, timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf(`grpc dial to host=%s: %w`, c.Host, err)
	}

	return NewFromGRPC(conn, log, timeout)
}

// NewFromGRPC allows to initialize peer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, log *zap.Logger, timeout time.Duration) (api.Peer, error) {
	l := log.Named(`NewFromGRPC`)
	p := &peer{conn: conn, log: log.Named(`peer`), timeout: timeout}
	if err := p.initEndorserClient(); err != nil {
		l.Error(`Failed to initialize endorser client`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to initialize EndorserClient`)
	}
	return p, nil
}
