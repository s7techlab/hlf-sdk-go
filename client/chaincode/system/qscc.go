package system

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	qscccore "github.com/hyperledger/fabric/core/scc/qscc"

	"github.com/s7techlab/hlf-sdk-go/api"
)

//go:embed qscc.swagger.json
var QSCCServiceSwagger []byte

type QSCCService struct {
	UnimplementedQSCCServiceServer
	Querier api.Querier
}

func NewQSCCService(querier api.Querier) *QSCCService {
	return &QSCCService{
		Querier: querier,
	}
}

func (q *QSCCService) ServiceDef() api.ServiceDef {
	return api.NewServiceDef(
		_QSCCService_serviceDesc.ServiceName,
		QSCCServiceSwagger,
		&_QSCCService_serviceDesc,
		q,
		RegisterQSCCServiceHandlerFromEndpoint,
	)
}

func (q *QSCCService) query(ctx context.Context, args []string, target proto.Message) (proto.Message, error) {
	var queryArgs [][]byte
	for _, arg := range args {
		queryArgs = append(queryArgs, []byte(arg))
	}

	res, err := q.Querier.Query(
		ctx, ``, QSCCName, queryArgs, nil, nil)

	if err != nil {
		return nil, fmt.Errorf(`query QSCC: %w`, err)
	}

	if err = proto.Unmarshal(res.Payload, target); err != nil {
		return nil, fmt.Errorf(`unmarshal result to %s: %w`, reflect.TypeOf(target), err)
	}

	return target, nil
}

func (q *QSCCService) GetChainInfo(ctx context.Context, request *GetChainInfoRequest) (*common.BlockchainInfo, error) {
	res, err := q.query(ctx,
		[]string{qscccore.GetChainInfo, request.ChannelName},
		&common.BlockchainInfo{})
	if err != nil {
		return nil, err
	}

	return res.(*common.BlockchainInfo), nil
}

func (q *QSCCService) GetBlockByNumber(ctx context.Context, request *GetBlockByNumberRequest) (*common.Block, error) {
	res, err := q.query(ctx,
		[]string{qscccore.GetBlockByNumber, request.ChannelName, strconv.FormatInt(request.BlockNumber, 10)},
		&common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (q *QSCCService) GetBlockByHash(ctx context.Context, request *GetBlockByHashRequest) (*common.Block, error) {
	res, err := q.query(ctx,
		[]string{qscccore.GetBlockByHash, request.ChannelName, string(request.BlockHash)},
		&common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (q *QSCCService) GetBlockByTxID(ctx context.Context, request *GetBlockByTxIDRequest) (*common.Block, error) {
	res, err := q.query(ctx,
		[]string{qscccore.GetBlockByTxID, request.ChannelName, request.TxId},
		&common.Block{})
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

func (q *QSCCService) GetTransactionByID(ctx context.Context, request *GetTransactionByIDRequest) (*peer.ProcessedTransaction, error) {
	res, err := q.query(ctx,
		[]string{qscccore.GetTransactionByID, request.ChannelName, request.TxId},
		&peer.ProcessedTransaction{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ProcessedTransaction), nil
}
