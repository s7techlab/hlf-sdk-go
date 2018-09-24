package peer

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	fabricPeer "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

type peer struct {
	log       *zap.Logger
	endpoints []string
	conn      *grpc.ClientConn
	connMx    sync.Mutex
	timeout   time.Duration
	client    fabricPeer.EndorserClient
}

func (p *peer) Endorse(ctx context.Context, proposal *fabricPeer.SignedProposal, opts ...api.PeerEndorseOpt) (*fabricPeer.ProposalResponse, error) {

	//TODO:  it all can be used WITHOUT THAT WRAPPER around EndorserClient
	if resp, err := p.client.ProcessProposal(ctx, proposal); err != nil {
		return nil, err
	} else {
		if resp.Response.Status != shim.OK {
			return nil, api.PeerEndorseError{Status: resp.Response.Status, Message: resp.Response.Message}
		}
		return resp, nil
	}
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
func New(c config.PeerConfig, log *zap.Logger) (api.Peer, error) {
	l := log.Named(`New`)
	conn, err := NewGRPCFromConfig(c, l)
	if err != nil {
		l.Debug(`Creating GRPC connection failed`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to initialize GRPC connection`)
	}
	l.Debug(`GRPC initialized`, zap.String(`target`, conn.Target()))
	return NewFromGRPC(conn, l)
}

// NewFromGRPC allows to initialize peer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, log *zap.Logger) (api.Peer, error) {
	l := log.Named(`NewFromGRPC`)
	p := &peer{conn: conn}
	if err := p.initEndorserClient(); err != nil {
		l.Debug(`Failed to initialize endorser client`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to initialize EndorserClient`)
	}
	return p, nil
}
