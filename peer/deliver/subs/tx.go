package subs

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"

	"github.com/s7techlab/hlf-sdk-go/api"
)

func NewTxSubscription(txId api.ChaincodeTx) *TxSubscription {
	return &TxSubscription{
		txId:   txId,
		result: make(chan *result, 1),
	}
}

type result struct {
	code peer.TxValidationCode
	err  error
}

type TxSubscription struct {
	txId   api.ChaincodeTx
	result chan *result
	ErrorCloser
}

func (ts *TxSubscription) Serve(sub ErrorCloser) *TxSubscription {
	ts.ErrorCloser = sub
	return ts
}

func (ts *TxSubscription) Result() (peer.TxValidationCode, error) {
	select {
	case r, ok := <-ts.result:
		if !ok {
			return -1, errors.New(`code is closed`)
		}
		return r.code, r.err
	case err, ok := <-ts.Err():
		if !ok {
			// NOTE: sometime error can be closed early thet result
			select {
			case r, ok := <-ts.result:
				if !ok {
					return -1, errors.New(`code is closed`)
				}
				return r.code, r.err
			default:
				return -1, errors.New(`err is closed`)
			}
		}
		return -1, err
	}
}

func (ts *TxSubscription) Handler(block *common.Block) bool {
	if block == nil {
		close(ts.result)
		return false
	}
	txFilter := util.TxValidationFlags(
		block.GetMetadata().GetMetadata()[common.BlockMetadataIndex_TRANSACTIONS_FILTER],
	)

	for i, r := range block.GetData().GetData() {
		env, err := utils.GetEnvelopeFromBlock(r)
		if err != nil {
			ts.result <- &result{code: 0, err: err}
			return true
		}

		p, err := utils.GetPayload(env)
		if err != nil {
			ts.result <- &result{code: 0, err: err}
			return true
		}

		chHeader, err := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
		if err != nil {
			ts.result <- &result{code: 0, err: err}
			return true
		}

		//println("TXID", chHeader.TxId, txFilter.IsValid(i))
		if api.ChaincodeTx(chHeader.TxId) == ts.txId {
			if txFilter.IsValid(i) {
				ts.result <- &result{code: txFilter.Flag(i), err: nil}
				return true
			} else {
				err = errors.Errorf("TxId validation code failed: %s", peer.TxValidationCode_name[int32(txFilter.Flag(i))])
				ts.result <- &result{code: txFilter.Flag(i), err: err}
				return true
			}
		}
	}

	return false
}
