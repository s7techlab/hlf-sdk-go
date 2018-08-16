package subs

import (
	"log"

	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"google.golang.org/grpc"
)

type txSubscription struct {
	txId      api.ChaincodeTx
	blockSub  api.BlockSubscription
	blockChan chan *common.Block
	events    chan api.TxEvent
}

func (ts *txSubscription) Result() (chan api.TxEvent, error) {
	var err error

	if ts.blockChan, err = ts.blockSub.Blocks(); err != nil {
		return nil, errors.Wrap(err, `failed to initiate block subscription`)
	}

	go ts.handleSubscription()

	return ts.events, nil
}

func (ts *txSubscription) handleSubscription() {
	for {
		select {
		case block, ok := <-ts.blockChan:
			if !ok {
				return
			}
			txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
			for i, r := range block.Data.Data {
				env, err := utils.GetEnvelopeFromBlock(r)
				if err != nil {
					ts.events <- api.TxEvent{
						TxId:    ts.txId,
						Success: false,
						Error:   err,
					}
					continue
				}

				p, err := utils.GetPayload(env)
				if err != nil {
					ts.events <- api.TxEvent{
						TxId:    ts.txId,
						Success: false,
						Error:   err,
					}
					continue
				}

				chHeader, err := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
				if err != nil {
					ts.events <- api.TxEvent{
						TxId:    ts.txId,
						Success: false,
						Error:   err,
					}
					continue
				}

				log.Println(ts.txId)

				if api.ChaincodeTx(chHeader.TxId) == ts.txId {
					if txFilter.IsValid(i) {
						ts.events <- api.TxEvent{
							TxId:    ts.txId,
							Success: true,
							Error:   nil,
						}
					} else {
						ts.events <- api.TxEvent{
							TxId:    ts.txId,
							Success: false,
							Error:   errors.Errorf("TxId validation code failed:%s", peer.TxValidationCode_name[int32(txFilter.Flag(i))]),
						}
					}
				}
			}
		case err, ok := <-ts.blockSub.Errors():
			if !ok {
				return
			}
			switch err.(type) {
			case *api.GRPCStreamError:
				ts.events <- api.TxEvent{TxId: ts.txId, Success: false, Error: err}
			default:
				ts.events <- api.TxEvent{TxId: ts.txId, Success: false, Error: errors.Wrap(err, `unknown error`)}
			}
		}
	}
}

func (ts *txSubscription) Close() error {
	return ts.blockSub.Close()
}

func NewTxSubscription(txId api.ChaincodeTx, channelName string, identity msp.SigningIdentity, conn *grpc.ClientConn, seekOpt ...api.EventCCSeekOption) api.TxSubscription {
	return &txSubscription{
		txId:      txId,
		blockSub:  NewBlockSubscription(channelName, identity, conn, seekOpt...),
		blockChan: make(chan *common.Block),
		events:    make(chan api.TxEvent),
	}
}
