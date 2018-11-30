package subs

import (
	"context"

	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"go.uber.org/zap"
)

type txSubscription struct {
	log  *zap.Logger
	txId api.ChaincodeTx

	blockChan chan *common.Block
	errChan   chan error

	ctx    context.Context
	cancel context.CancelFunc
}

func (ts *txSubscription) Result() (peer.TxValidationCode, error) {
	log := ts.log.Named(`Result`)

	log.Debug(`Reading blockChan`)
	for {
		select {
		case block, ok := <-ts.blockChan:
			if !ok {
				log.Debug(`blockChan is closed`)
				return -1, errors.New(`blockChan is closed`)
			}
			log.Debug(`Received block`, zap.Uint64(`number`, block.Header.Number),
				zap.ByteString(`hash`, block.Header.DataHash),
				zap.ByteString(`prevHash`, block.Header.PreviousHash),
			)
			if outBlock, code, err := ts.handleBlock(block); !outBlock {
				if err != nil {
					log.Error(`Block parse error`, zap.Error(err))
				}
				continue
			} else {
				return code, err
			}
		case <-ts.ctx.Done():
			log.Debug(`Context canceled`, zap.Error(ts.ctx.Err()))
			return -1, ts.ctx.Err()
		}
	}
}

func (ts *txSubscription) handleBlock(block *common.Block) (bool, peer.TxValidationCode, error) {
	log := ts.log.Named(`handleBlock`)
	txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	log.Debug(`Got txFilter`, zap.Uint8s(`txFilter`, txFilter))

	for i, r := range block.Data.Data {
		log.Debug(`Parsing common.Envelope`)
		env, err := utils.GetEnvelopeFromBlock(r)
		if err != nil {
			log.Error(`Parse common.Envelope error`, zap.Error(err))
			return false, 0, err
		}
		log.Debug(`Parsed common.Envelope`, zap.ByteString(`payload`, env.Payload), zap.ByteString(`signature`, env.Signature))

		log.Debug(`Parsing common.Payload`)
		p, err := utils.GetPayload(env)
		if err != nil {
			log.Error(`Parse common.Payload error`, zap.Error(err))
			return false, 0, err
		}
		log.Debug(`Parsed common.Payload`)

		log.Debug(`Parsing common.ChannelHeader`)
		chHeader, err := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
		if err != nil {
			log.Error(`Parse common.ChannelHeader failed`, zap.Error(err))
			return false, 0, err
		}
		log.Debug(`Parsed common.ChannelHeader`)

		log.Debug(`Comparing txId`, zap.String(`txId`, chHeader.TxId), zap.String(`searchTxId`, string(ts.txId)))
		if api.ChaincodeTx(chHeader.TxId) == ts.txId {
			log.Debug(`Check transaction`, zap.String(`txId`, chHeader.TxId))
			if txFilter.IsValid(i) {
				log.Debug(`Transaction is valid`, zap.String(`txId`, chHeader.TxId))
				return true, txFilter.Flag(i), nil
			} else {
				err = errors.Errorf("TxId validation code failed: %s", peer.TxValidationCode_name[int32(txFilter.Flag(i))])
				log.Debug(`Transaction is invalid`, zap.Error(err))
				return true, txFilter.Flag(i), err
			}
		}
	}

	return false, 0, nil
}

func (ts *txSubscription) Close() error {
	ts.log.Named(`Close`).Debug(`Closing subscription`)
	ts.cancel()
	return nil
}

func NewTxSubscription(ctx context.Context, txId api.ChaincodeTx, blockChan chan *common.Block, errChan chan error, log *zap.Logger) api.TxSubscription {
	l := log.Named(`TxSubscription`)

	newCtx, cancel := context.WithCancel(ctx)
	return &txSubscription{
		log:       l,
		txId:      txId,
		blockChan: blockChan,
		errChan:   errChan,
		ctx:       newCtx,
		cancel:    cancel,
	}
}
