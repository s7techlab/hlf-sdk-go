package qscc

import (
	"context"
	_ "embed"
	"strconv"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	qscccore "github.com/hyperledger/fabric/core/scc/qscc"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/proto/systemcc/qscc"
	"github.com/s7techlab/hlf-sdk-go/service"
)

type QSCCService struct {
	qscc.UnimplementedQSCCServiceServer
	Querier *tx.ProtoQuerier
}

func NewQSCC(querier api.Querier) *QSCCService {
	return &QSCCService{
		Querier: tx.NewProtoQuerier(querier, ``, chaincode.QSCC),
	}
}

func (q *QSCCService) ServiceDef() *service.Def {
	return service.NewDef(
		qscc.ServiceDesc.ServiceName,
		qscc.Swagger,
		&qscc.ServiceDesc,
		q,
		qscc.RegisterQSCCServiceHandlerFromEndpoint,
	)
}

func (q *QSCCService) GetChainInfo(ctx context.Context, request *qscc.GetChainInfoRequest) (*common.BlockchainInfo, error) {
	res, err := q.Querier.QueryStringsProto(ctx, []string{qscccore.GetChainInfo, request.ChannelName}, &common.BlockchainInfo{})
	if err != nil {
		return nil, err
	}
	return res.(*common.BlockchainInfo), nil
}

func (q *QSCCService) GetBlockByNumber(ctx context.Context, request *qscc.GetBlockByNumberRequest) (*common.Block, error) {
	res, err := q.Querier.QueryStringsProto(ctx,
		[]string{qscccore.GetBlockByNumber, request.ChannelName, strconv.FormatInt(request.BlockNumber, 10)},
		&common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (q *QSCCService) GetBlockByHash(ctx context.Context, request *qscc.GetBlockByHashRequest) (*common.Block, error) {
	res, err := q.Querier.QueryStringsProto(ctx,
		[]string{qscccore.GetBlockByHash, request.ChannelName, string(request.BlockHash)},
		&common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (q *QSCCService) GetBlockByTxID(ctx context.Context, request *qscc.GetBlockByTxIDRequest) (*common.Block, error) {
	res, err := q.Querier.QueryStringsProto(ctx,
		[]string{qscccore.GetBlockByTxID, request.ChannelName, request.TxId},
		&common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (q *QSCCService) GetTransactionByID(ctx context.Context, request *qscc.GetTransactionByIDRequest) (
	*peer.ProcessedTransaction, error) {
	res, err := q.Querier.QueryStringsProto(ctx,
		[]string{qscccore.GetTransactionByID, request.ChannelName, request.TxId},
		&peer.ProcessedTransaction{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ProcessedTransaction), nil
}
