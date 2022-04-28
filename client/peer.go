package client

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
	grpcclient "github.com/s7techlab/hlf-sdk-go/client/grpc"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/peer/deliver"
)

const (
	// FnShowMaxLength limit to show fn name (args[0]) in debug and error messages
	FnShowMaxLength = 100

	PeerDefaultDialTimeout    = 5 * time.Second
	PeerDefaultEndorseTimeout = 5 * time.Second
)

type peer struct {
	conn     *grpc.ClientConn
	client   fabricPeer.EndorserClient
	identity msp.SigningIdentity

	endorseDefaultTimeout time.Duration

	logger *zap.Logger
}

// NewPeer returns new peer instance bassed on peer config
func NewPeer(dialCtx context.Context, c config.ConnectionConfig, identity msp.SigningIdentity, logger *zap.Logger) (api.Peer, error) {
	opts, err := grpcclient.OptionsFromConfig(c, logger)
	if err != nil {
		return nil, fmt.Errorf(`peer grpc options from config: %w`, err)
	}

	dialTimeout := c.Timeout.Duration
	if dialTimeout == 0 {
		dialTimeout = PeerDefaultDialTimeout
	}

	// Dial shoould always has timeout
	ctxDeadline, exists := dialCtx.Deadline()
	if !exists {
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(dialCtx, dialTimeout)
		defer cancel()

		ctxDeadline, _ = dialCtx.Deadline()
	}

	logger.Debug(`dial to peer`,
		zap.String(`host`, c.Host), zap.Time(`context deadline`, ctxDeadline))
	conn, dialErr := grpc.DialContext(dialCtx, c.Host, opts...)
	if dialErr != nil {
		return nil, fmt.Errorf(`grpc dial to peer endpoint=%s: %w`, c.Host, err)
	}

	return NewFromGRPC(conn, identity, logger, c.Timeout.Duration)
}

// NewFromGRPC allows initializing peer from existing GRPC connection
func NewFromGRPC(conn *grpc.ClientConn, identity msp.SigningIdentity, logger *zap.Logger, endorseDefaultTimeout time.Duration) (api.Peer, error) {
	if conn == nil {
		return nil, errors.New(`empty connection`)
	}

	if endorseDefaultTimeout == 0 {
		endorseDefaultTimeout = PeerDefaultEndorseTimeout
	}

	p := &peer{
		conn:                  conn,
		client:                fabricPeer.NewEndorserClient(conn),
		identity:              identity,
		endorseDefaultTimeout: endorseDefaultTimeout,
		logger:                logger.Named(`peer`),
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

	signer = tx.ChooseSigner(ctx, signer, p.identity)

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
	if _, ok := ctx.Deadline(); !ok && p.endorseDefaultTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.endorseDefaultTimeout)
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
	if identity == nil {
		identity = p.identity
	}
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
