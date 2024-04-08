package client

import (
	"context"
	"fmt"
	"sync"
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
	"github.com/s7techlab/hlf-sdk-go/block"
	"github.com/s7techlab/hlf-sdk-go/client/channel"
	"github.com/s7techlab/hlf-sdk-go/client/deliver"
	grpcclient "github.com/s7techlab/hlf-sdk-go/client/grpc"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/service/systemcc/qscc"
)

const (
	// FnShowMaxLength limit to show fn name (args[0]) in debug and error messages
	FnShowMaxLength = 100

	PeerDefaultDialTimeout    = 5 * time.Second
	PeerDefaultEndorseTimeout = 5 * time.Second

	DefaultPeerChannelsObservePeriod = 30 * time.Second
)

type peer struct {
	conn        *grpc.ClientConn
	client      fabricPeer.EndorserClient
	identity    msp.SigningIdentity
	tlsCertHash []byte

	endorseDefaultTimeout time.Duration

	configBlocks map[string]*common.Block
	mu           sync.Mutex

	logger *zap.Logger
}

// NewPeer returns new peer instance based on peer config
func NewPeer(ctx context.Context, c config.ConnectionConfig, identity msp.SigningIdentity, logger *zap.Logger) (api.Peer, error) {
	opts, err := grpcclient.OptionsFromConfig(c, logger)
	if err != nil {
		return nil, fmt.Errorf(`peer grpc options from config: %w`, err)
	}

	dialTimeout := c.Timeout.Duration
	if dialTimeout == 0 {
		dialTimeout = PeerDefaultDialTimeout
	}

	// Dial should always have timeout
	var dialCtx context.Context
	ctxDeadline, exists := ctx.Deadline()
	if !exists {
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(ctx, dialTimeout)
		defer cancel()

		ctxDeadline, _ = dialCtx.Deadline()
	}

	logger.Debug(`dial to peer`, zap.String(`host`, c.Host), zap.Time(`context deadline`, ctxDeadline))
	conn, err := grpc.DialContext(dialCtx, c.Host, opts.Dial...)
	if err != nil {
		return nil, fmt.Errorf(`grpc dial to peer endpoint=%s: %w`, c.Host, err)
	}

	return NewFromGRPC(ctx, conn, identity, opts.TLSCertHash, logger, c.Timeout.Duration)
}

// NewFromGRPC allows initializing peer from existing GRPC connection
func NewFromGRPC(ctx context.Context, conn *grpc.ClientConn, identity msp.SigningIdentity, tlsCertHash []byte, logger *zap.Logger, endorseDefaultTimeout time.Duration) (api.Peer, error) {
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
		tlsCertHash:           tlsCertHash,
		endorseDefaultTimeout: endorseDefaultTimeout,
		configBlocks:          make(map[string]*common.Block),
		logger:                logger.Named(`peer`),
	}

	qsccService := qscc.NewQSCC(p)

	if err := p.getConfigBlocks(ctx, qsccService); err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(DefaultPeerChannelsObservePeriod)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := p.getConfigBlocks(ctx, qsccService); err != nil {
					p.logger.Error("get config blocks", zap.Error(err))
				}
			}
		}
	}()

	return p, nil
}

func (p *peer) getConfigBlocks(ctx context.Context, qsccService *qscc.QSCCService) error {
	channels, err := p.GetChannels(ctx)
	if err != nil {
		return fmt.Errorf("get all channels: %w", err)
	}

	if len(channels.Channels) <= len(p.configBlocks) {
		return nil
	}

	for _, ch := range channels.GetChannels() {
		_, exist := p.configBlocks[ch.ChannelId]
		if !exist {
			configBlock, err := qsccService.GetBlockByNumber(ctx, &qscc.GetBlockByNumberRequest{ChannelName: ch.ChannelId, BlockNumber: 0})
			if err != nil {
				return fmt.Errorf("get block by number from channel %s: %w", ch.ChannelId, err)
			}

			if configBlock != nil {
				p.configBlocks[ch.ChannelId] = configBlock
			}
		}
	}

	return nil
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

func (p *peer) Blocks(ctx context.Context, channel string, identity msp.SigningIdentity, blockRange ...int64) (<-chan *common.Block, func() error, error) {
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

func (p *peer) ParsedBlocks(ctx context.Context, channel string, identity msp.SigningIdentity, blockRange ...int64) (<-chan *block.Block, func() error, error) {
	commonBlocks, commonCloser, err := p.Blocks(ctx, channel, identity, blockRange...)
	if err != nil {
		return nil, nil, err
	}

	parsedBlockChan := make(chan *block.Block)
	go func() {
		defer func() {
			close(parsedBlockChan)
		}()

		for {
			select {
			case b, ok := <-commonBlocks:
				if !ok {
					return
				}
				if b == nil {
					return
				}

				p.mu.Lock()
				configBlock := p.configBlocks[channel]
				p.mu.Unlock()

				parsedBlock, err := block.ParseBlock(b, block.WithConfigBlock(configBlock))
				if err != nil {
					p.logger.Error("parse block", zap.String("channel", channel), zap.Uint64("number", b.Header.Number))
					continue
				}

				parsedBlockChan <- parsedBlock
			}
		}
	}()

	parsedCloser := func() error {
		if closerErr := commonCloser(); closerErr != nil {
			return closerErr
		}
		return nil
	}

	return parsedBlockChan, parsedCloser, nil
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
	return qscc.NewQSCC(p).GetChainInfo(ctx, &qscc.GetChainInfoRequest{ChannelName: channel})
}

func (p *peer) GetChannels(ctx context.Context) (*fabricPeer.ChannelQueryResponse, error) {
	return channel.NewCSCCListGetter(p).GetChannels(ctx)
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
	return deliver.New(fabricPeer.NewDeliverClient(p.conn), identity, p.tlsCertHash), nil
}

// CurrentIdentity defaultSigner returns current signing defaultSigner used by Client
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
