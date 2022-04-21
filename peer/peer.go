package peer

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver"
	"github.com/s7techlab/hlf-sdk-go/util"
)

type peer struct {
	conn     *grpc.ClientConn
	timeout  time.Duration
	client   fabricPeer.EndorserClient
	identity msp.SigningIdentity
	logger   *zap.Logger
}

var (
	defaultTimeout = 5 * time.Second
)

// New returns new peer instance based on peer config
func New(c config.ConnectionConfig, identity msp.SigningIdentity, logger *zap.Logger) (api.Peer, error) {
	opts, err := util.NewGRPCOptionsFromConfig(c, logger)
	if err != nil {
		return nil, fmt.Errorf(`grpc options from config: %w`, err)
	}

	timeout := c.Timeout.Duration
	if timeout == 0 {
		timeout = defaultTimeout
	}

	//ctx, _ := context.WithTimeout(context.Background(), c.Timeout.Duration)
	logger.Debug(`dial to peer`, zap.String(`host`, c.Host), zap.Duration(`timeout`, timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf(`grpc dial to host=%s: %w`, c.Host, err)
	}

	return NewFromGRPC(conn, identity, logger, timeout)
}

// NewFromGRPC allows initializing peer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, identity msp.SigningIdentity, logger *zap.Logger, timeout time.Duration) (api.Peer, error) {
	if conn == nil {
		return nil, errors.New(`empty connection`)
	}

	p := &peer{
		conn:     conn,
		client:   fabricPeer.NewEndorserClient(conn),
		identity: identity,
		logger:   logger.Named(`peer`),
		timeout:  timeout,
	}

	return p, nil
}

func (p *peer) Query(
	ctx context.Context,
	channel string,
	chaincode string,
	args [][]byte,
	signer msp.SigningIdentity,
	transientMap map[string][]byte) (*fabricPeer.Response, error) {

	p.logger.Debug(`endorser query`,
		zap.String(`channel`, channel),
		zap.String(`chaincode`, chaincode),
		zap.String(`args[0] (fn)`, string(args[0])))

	if signer == nil {
		signer = p.identity
	}

	proposal, _, err := tx.Endorsement{
		Channel:      channel,
		Chaincode:    chaincode,
		Args:         args,
		Signer:       signer,
		TransientMap: transientMap,
	}.SignedProposal()

	if err != nil {
		return nil, err
	}

	response, err := p.Endorse(ctx, proposal)
	if err != nil {
		return nil, err
	}

	return response.Response, nil
}

func (p *peer) Endorse(ctx context.Context, proposal *fabricPeer.SignedProposal) (*fabricPeer.ProposalResponse, error) {
	if _, ok := ctx.Deadline(); !ok && p.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	p.logger.Debug(`endorse`, zap.String(`uri`, p.Uri()))

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
