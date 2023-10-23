package channel

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
)

type CSCCListGetter struct {
	Querier *tx.ProtoQuerier
}

func NewCSCCListGetter(querier api.Querier) *CSCCListGetter {
	return &CSCCListGetter{
		Querier: tx.NewProtoQuerier(querier, ``, chaincode.CSCC),
	}
}

func (g *CSCCListGetter) GetChannels(ctx context.Context) (*peer.ChannelQueryResponse, error) {
	res, err := g.Querier.QueryStringsProto(ctx, []string{chaincode.CSCCGetChannels}, &peer.ChannelQueryResponse{})
	if err != nil {
		return nil, err
	}
	return res.(*peer.ChannelQueryResponse), nil
}
