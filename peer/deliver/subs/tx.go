package subs

import (
	"context"

	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type txSubscription struct {
	log       *zap.Logger
	txId      api.ChaincodeTx
	blockSub  api.BlockSubscription
	blockChan chan *common.Block
	events    chan api.TxEvent
}

func (ts *txSubscription) Result() (chan api.TxEvent, error) {
	log := ts.log.Named(`Result`)

	var err error

	log.Debug(`Initializing blockSubscription`)
	if ts.blockChan, err = ts.blockSub.Blocks(); err != nil {
		log.Error(`Failed to initiate block subscription`, zap.Error(err))
		return nil, errors.Wrap(err, `failed to initiate block subscription`)
	}

	log.Debug(`Starting handleSubscription`)
	go ts.handleSubscription()

	return ts.events, nil
}

func (ts *txSubscription) handleSubscription() {
	log := ts.log.Named(`handleSubscription`)
	for {
		select {
		case block, ok := <-ts.blockChan:
			if !ok {
				log.Debug(`blockChan is closed`)
				return
			}

			log.Debug(`Received block`, zap.Uint64(`number`, block.Header.Number),
				zap.ByteString(`hash`, block.Header.DataHash),
				zap.ByteString(`prevHash`, block.Header.PreviousHash),
			)

			txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
			log.Debug(`Got txFilter`, zap.Uint8s(`txFilter`, txFilter))

			for i, r := range block.Data.Data {
				log.Debug(`Parsing common.Envelope`)
				env, err := utils.GetEnvelopeFromBlock(r)
				if err != nil {
					log.Error(`Parse common.Envelope error`, zap.Error(err))
					ts.events <- api.TxEvent{
						TxId:    ts.txId,
						Success: false,
						Error:   err,
					}
					continue
				}
				log.Debug(`Parsed common.Envelope`, zap.ByteString(`payload`, env.Payload), zap.ByteString(`signature`, env.Signature))

				log.Debug(`Parsing common.Payload`)
				p, err := utils.GetPayload(env)
				if err != nil {
					log.Error(`Parse common.Payload error`, zap.Error(err))
					ts.events <- api.TxEvent{
						TxId:    ts.txId,
						Success: false,
						Error:   err,
					}
					continue
				}
				log.Debug(`Parsed common.Payload`)

				log.Debug(`Parsing common.ChannelHeader`)
				chHeader, err := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
				if err != nil {

					ts.events <- api.TxEvent{
						TxId:    ts.txId,
						Success: false,
						Error:   err,
					}
					continue
				}
				log.Debug(`Parsed common.ChannelHeader`)

				log.Debug(`Comparing txId`, zap.String(`txId`, chHeader.TxId), zap.String(`searchTxId`, string(ts.txId)))
				if api.ChaincodeTx(chHeader.TxId) == ts.txId {
					log.Debug(`Check transaction`, zap.String(`txId`, chHeader.TxId))
					if txFilter.IsValid(i) {
						log.Debug(`Sending api.TxEvent`, zap.String(`txId`, chHeader.TxId), zap.Bool(`success`, true), zap.Error(nil))
						ts.events <- api.TxEvent{
							TxId:    ts.txId,
							Success: true,
							Error:   nil,
						}
						log.Debug(`Sent api.TxEvent`)
					} else {

						err = errors.Errorf("TxId validation code failed:%s", peer.TxValidationCode_name[int32(txFilter.Flag(i))])
						log.Debug(`Sending api.TxEvent`, zap.String(`txId`, chHeader.TxId), zap.Bool(`success`, false), zap.Error(err))
						ts.events <- api.TxEvent{
							TxId:    ts.txId,
							Success: false,
							Error:   err,
						}
						log.Debug(`Send api.txEvent`)
					}
				}
			}
		case err, ok := <-ts.blockSub.Errors():
			log.Debug(`Reading blockSub.Errors`)
			if !ok {
				log.Debug(`blockSub.Errors is closed`)
				return
			}
			log.Error(`Got blockSub error`, zap.Error(err))
			switch err.(type) {
			case *api.GRPCStreamError:
				log.Debug(`Sending api.TxEvent`, zap.String(`txId`, string(ts.txId)), zap.Bool(`success`, false), zap.Error(err))
				ts.events <- api.TxEvent{TxId: ts.txId, Success: false, Error: err}
				log.Debug(`Send api.txEvent`)
			default:
				err = errors.Wrap(err, `unknown error`)
				log.Debug(`Sending api.TxEvent`, zap.String(`txId`, string(ts.txId)), zap.Bool(`success`, false), zap.Error(err))
				ts.events <- api.TxEvent{TxId: ts.txId, Success: false, Error: err}
				log.Debug(`Send api.txEvent`)
			}
		}
	}
}

func (ts *txSubscription) Close() error {
	ts.log.Named(`Close`).Debug(`Closing subscription`)
	return ts.blockSub.Close()
}

func NewTxSubscription(ctx context.Context, txId api.ChaincodeTx, channelName string, identity msp.SigningIdentity, conn *grpc.ClientConn, log *zap.Logger, seekOpt ...api.EventCCSeekOption) api.TxSubscription {
	l := log.Named(`TxSubscription`)
	return &txSubscription{
		log:       l,
		txId:      txId,
		blockSub:  NewBlockSubscription(ctx, channelName, identity, conn, l, seekOpt...),
		blockChan: make(chan *common.Block),
		events:    make(chan api.TxEvent),
	}
}
