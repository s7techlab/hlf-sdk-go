package peer

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/common"
	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver"
	"github.com/s7techlab/hlf-sdk-go/util"
)

// FnShowMaxLength limit to show fn name (args[0]) in debug and error messages
const FnShowMaxLength = 100

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

	fn := ``
	zapFields := []zap.Field{
		zap.String(`channel`, channel),
		zap.String(`chaincode`, chaincode),
	}

	if len(args) > 0 {
		if len(args[0]) < FnShowMaxLength {
			fn = string(args[0])
			zapFields = append(zapFields, zap.String(`args[0] (fn)`, fn))
		}
	}

	if signer == nil && p.identity != nil {
		signer = p.identity
		zapFields = append(zapFields, zap.String(`use default identity`, p.identity.GetMSPIdentifier()))
	}

	p.logger.Debug(`peer query`, zapFields...)

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
		return nil, fmt.Errorf(`peer query channel=%s chaincode=%s fn=%s: %w`, channel, chaincode, fn, err)
	}

	return response.Response, nil
}

func (p *peer) Blocks(ctx context.Context, channel string, identity msp.SigningIdentity, blockRange ...int64) (blockChan <-chan *common.Block, closer func() error, err error) {
	p.logger.Debug(`peer blocks request`,
		zap.String(`uri`, p.Uri()),
		zap.String(`channel`, channel),
		zap.Reflect(`range`, blockRange))

	dc, err := p.DeliverClient(identity)
	if err != nil {
		return nil, nil, fmt.Errorf(`deliver client: %w`, err)
	}

	var seekOpts []api.EventCCSeekOption
	seekOpt, err := deliver.NewSeekOptConverter(p, p.logger).ByBlockRange(ctx, channel, blockRange...)
	if err != nil {
		return nil, nil, err
	}

	if seekOpt != nil {
		seekOpts = append(seekOpts, seekOpt)
	}

	bs, err := dc.SubscribeBlock(ctx, channel, seekOpts...)
	if err != nil {
		return nil, nil, err
	}

	return bs.Blocks(), bs.Close, nil
}

func (p *peer) Events(ctx context.Context, channel string, chaincode string, identity msp.SigningIdentity, blockRange ...int64) (events chan interface {
	Event() *fabricPeer.ChaincodeEvent
	Block() uint64
	TxTimestamp() *timestamp.Timestamp
}, closer func() error, err error) {

	p.logger.Debug(`peer events request`,
		zap.String(`uri`, p.Uri()),
		zap.String(`channel`, channel),
		zap.Reflect(`range`, blockRange))

	dc, err := p.DeliverClient(identity)
	if err != nil {
		return nil, nil, fmt.Errorf(`deliver client: %w`, err)
	}
	var seekOpts []api.EventCCSeekOption
	seekOpt, err := deliver.NewSeekOptConverter(p, p.logger).ByBlockRange(ctx, channel, blockRange...)
	if err != nil {
		return nil, nil, err
	}

	if seekOpt != nil {
		seekOpts = append(seekOpts, seekOpt)
	}

	sub, err := dc.SubscribeCC(ctx, channel, chaincode, seekOpts...)
	if err != nil {
		return nil, nil, err
	}

	return sub.EventsExtended(), sub.Close, nil
}

func (p *peer) GetChainInfo(ctx context.Context, channel string) (*common.BlockchainInfo, error) {
	return system.NewQSCC(p).GetChainInfo(ctx, &system.GetChainInfoRequest{ChannelName: channel})
}

func (p *peer) GetChannels(ctx context.Context) (*fabricPeer.ChannelQueryResponse, error) {
	return system.NewCSCCChannelsFetcher(p).GetChannels(ctx)
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

// CurrentIdentity identity returns current signing identity used by core
func (p *peer) CurrentIdentity() msp.SigningIdentity {
	return p.identity
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
