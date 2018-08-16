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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

type peer struct {
	uri         string
	conn        *grpc.ClientConn
	ctx         context.Context
	connMx      sync.Mutex
	grpcOptions []grpc.DialOption
	timeout     time.Duration
	client      fabricPeer.EndorserClient
}

func (p *peer) Endorse(proposal *fabricPeer.SignedProposal, opts ...api.PeerEndorseOpt) (*fabricPeer.ProposalResponse, error) {

	var err error

	endorseOpts := new(api.PeerEndorseOpts)
	for _, opt := range opts {
		if err = opt(endorseOpts); err != nil {
			return nil, errors.Wrap(err, `failed to apply option`)
		}
	}

	// Use default peer context if endorse context isn't provided
	if endorseOpts.Context == nil {
		endorseOpts.Context = p.ctx
	}

	if resp, err := p.client.ProcessProposal(endorseOpts.Context, proposal); err != nil {
		return nil, errors.Wrap(err, `failed to endorse proposal`)
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
	return p.uri
}

func (p *peer) Close() error {
	return p.conn.Close()
}

func (p *peer) initEndorserClient() error {
	var err error
	if p.conn == nil {
		p.connMx.Lock()
		defer p.connMx.Unlock()
		if p.conn, err = grpc.DialContext(p.ctx, p.uri, p.grpcOptions...); err != nil {
			return errors.Wrap(err, `failed to initialize grpc connection`)
		}
	}

	if p.client == nil {
		p.client = fabricPeer.NewEndorserClient(p.conn)
	}

	return nil
}

// New returns new peer instance based on peer config
func New(c config.PeerConfig) (api.Peer, error) {
	p := peer{uri: c.Host, grpcOptions: make([]grpc.DialOption, 0)}
	if c.Tls.Enabled {
		if ts, err := credentials.NewClientTLSFromFile(c.Tls.CertPath, ``); err != nil {
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			p.grpcOptions = append(p.grpcOptions, grpc.WithTransportCredentials(ts))
		}
	} else {
		p.grpcOptions = append(p.grpcOptions, grpc.WithInsecure())
	}

	if c.Timeout.Duration != 0 {
		p.ctx, _ = context.WithTimeout(context.Background(), c.Timeout.Duration)
	} else {
		p.ctx = context.Background()
	}

	// Set KeepAlive parameters if presented
	if c.GRPC.KeepAlive != nil {
		p.grpcOptions = append(p.grpcOptions, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Duration(c.GRPC.KeepAlive.Time) * time.Second,
			Timeout: time.Duration(c.GRPC.KeepAlive.Timeout) * time.Second,
		}))
	}

	p.grpcOptions = append(p.grpcOptions, grpc.WithBlock(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	if err := p.initEndorserClient(); err != nil {
		return nil, errors.Wrap(err, `failed to initialize EndorserClient`)
	}
	return &p, nil
}

// NewFromGRPC allows to initialize peer from existing GRPC connection
func NewFromGRPC(address string, conn *grpc.ClientConn, grpcOptions ...grpc.DialOption) (api.Peer, error) {
	p := peer{conn: conn, uri: address, grpcOptions: grpcOptions}
	if err := p.initEndorserClient(); err != nil {
		return nil, errors.Wrap(err, `failed to initialize EndorserClient`)
	}
	return &p, nil
}
