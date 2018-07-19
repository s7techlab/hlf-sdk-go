package peer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	commonUtil "github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	fabricPeer "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// TODO implement logger instead of log package
type eventHub struct {
	uri      string
	opts     []grpc.DialOption
	identity msp.SigningIdentity
	conn     *grpc.ClientConn
	connMx   sync.Mutex
}

func (e *eventHub) getSubsCode(channelName string, ccName string) string {
	return fmt.Sprintf("%s_%s", channelName, ccName)
}

func (e *eventHub) SubscribeCC(channelName string, ccName string, seekOpt ...api.EventCCSeekOption) api.EventCCSubscription {
	sub := &eventSubscription{
		ccName:      ccName,
		channelName: channelName,
		errChan:     make(chan error),
		events:      make(chan *fabricPeer.ChaincodeEvent),
		parentConn:  e.conn,
		identity:    e.identity,
	}
	if len(seekOpt) > 0 {
		sub.startPos, sub.stopPos = seekOpt[0]()
	} else {
		sub.startPos, sub.stopPos = api.SeekNewest()()
	}
	return sub
}

func (e *eventHub) SubscribeTx(channelName string, txId api.ChaincodeTx) api.EventTxSubscription {
	sub := &txSubscription{
		txId:        txId,
		channelName: channelName,
		events:      make(chan api.TxEvent),
		parentConn:  e.conn,
		identity:    e.identity,
	}
	return sub
}

func (e *eventHub) Close() error {
	return e.conn.Close()
}

func (e *eventHub) initConnection() error {
	var err error
	e.connMx.Lock()
	defer e.connMx.Unlock()
	if e.conn == nil || e.conn.GetState() == connectivity.Shutdown {
		if e.conn, err = grpc.Dial(e.uri, e.opts...); err != nil {
			return errors.Wrap(err, `failed to initialize grpc connection`)
		}
	}
	return nil
}

func getCCEventFromEnvelope(envelopeData []byte) (*fabricPeer.ChaincodeEvent, error) {
	if envelopeData == nil {
		return nil, errors.New(`no envelope data`)
	}
	if envelope, err := utils.GetEnvelopeFromBlock(envelopeData); err != nil {
		return nil, errors.Wrap(err, `failed to get envelope`)
	} else {
		if payload, err := utils.GetPayload(envelope); err != nil {
			return nil, errors.Wrap(err, `failed to get payload from envelope`)
		} else {
			if channelHeader, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader); err != nil {
				return nil, errors.Wrap(err, `failed to unmarshal channel header`)
			} else {
				switch common.HeaderType(channelHeader.Type) {
				case common.HeaderType_ENDORSER_TRANSACTION:
					if tx, err := utils.GetTransaction(payload.Data); err != nil {
						return nil, errors.Wrap(err, `failed to get transaction`)
					} else {
						if ccActionPayload, err := utils.GetChaincodeActionPayload(tx.Actions[0].Payload); err != nil {
							return nil, errors.Wrap(err, `failed to get chaincode action payload`)
						} else {
							if propRespPayload, err := utils.GetProposalResponsePayload(ccActionPayload.Action.ProposalResponsePayload); err != nil {
								return nil, errors.Wrap(err, `failed to get proposal response payload`)
							} else {
								if caPayload, err := utils.GetChaincodeAction(propRespPayload.Extension); err != nil {
									return nil, errors.Wrap(err, `failed to get chaincode action`)
								} else {
									if ccEvent, err := utils.GetChaincodeEvents(caPayload.Events); err != nil {
										return nil, errors.Wrap(err, `failed to get events`)
									} else {
										return ccEvent, nil
									}
								}
							}
						}
					}
				default:
					return nil, errors.Errorf("err unknown tx type: %s", common.HeaderType_name[channelHeader.Type])
				}
			}
		}
	}
}

func NewEventHub(config config.PeerConfig, identity msp.SigningIdentity, grpcOptions ...grpc.DialOption) (api.EventHub, error) {
	var err error
	evHub := &eventHub{
		uri:      config.Host,
		opts:     grpcOptions,
		identity: identity,
	}

	if config.Tls.Enabled {
		if ts, err := credentials.NewClientTLSFromFile(config.Tls.CertPath, ``); err != nil {
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			evHub.opts = append(evHub.opts, grpc.WithTransportCredentials(ts))
		}
	} else {
		evHub.opts = append(evHub.opts, grpc.WithInsecure())
	}

	// Set KeepAlive parameters if presented
	if config.GRPC.KeepAlive != nil {
		evHub.opts = append(evHub.opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Duration(config.GRPC.KeepAlive.Time) * time.Second,
			Timeout: time.Duration(config.GRPC.KeepAlive.Timeout) * time.Second,
		}))
	}

	evHub.opts = append(evHub.opts, grpc.WithBlock(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	if err = evHub.initConnection(); err != nil {
		return nil, errors.Wrap(err, `failed to initialize EventHub`)
	}

	return evHub, nil
}

// GRPCStreamError contains original error from GRPC stream
type GRPCStreamError struct {
	Err error
}

func (e *GRPCStreamError) Error() string {
	return fmt.Sprintf("grpc stream error: %s", e.Err)
}

type EnvelopeParsingError struct {
	Err error
}

func (e *EnvelopeParsingError) Error() string {
	return fmt.Sprintf("envelope parsing error: %s", e.Err)
}

type UnknownEventTypeError struct {
	Type string
}

func (e *UnknownEventTypeError) Error() string {
	return fmt.Sprintf("unknown event type: %s", e.Type)
}

type InvalidTxError struct {
	TxId api.ChaincodeTx
	Code fabricPeer.TxValidationCode
}

func (e *InvalidTxError) Error() string {
	return fmt.Sprintf("invalid tx: %s with validation code: %s", e.TxId, e.Code.String())
}

type eventSubscription struct {
	startPos    *orderer.SeekPosition
	stopPos     *orderer.SeekPosition
	client      fabricPeer.Deliver_DeliverClient
	ccName      string
	channelName string
	events      chan *fabricPeer.ChaincodeEvent
	errChan     chan error
	parentConn  *grpc.ClientConn
	identity    msp.SigningIdentity
}

func (es *eventSubscription) Events() (chan *fabricPeer.ChaincodeEvent, error) {
	var err error
	if es.client, err = fabricPeer.NewDeliverClient(es.parentConn).Deliver(context.Background()); err != nil {
		return nil, errors.Wrap(err, `failed to get deliver client`)
	}

	if env, err := seekEnvelope(es.channelName, es.startPos, es.stopPos, es.identity); err != nil {
		return nil, errors.Wrap(err, `failed to get seek envelope`)
	} else {
		if err = es.client.Send(env); err != nil {
			return nil, errors.Wrap(err, `failed to send seek envelope`)
		} else {
			//if _, err := es.client.Recv(); err != nil {
			//	return nil, errors.Wrap(err, `failed to get deliver response`)
			//}
		}
	}

	go es.handleCCSubscription()

	return es.events, nil
}

func (es *eventSubscription) Errors() (chan error) {
	return es.errChan
}

func (es *eventSubscription) handleCCSubscription() {
	for {
		ev, err := es.client.Recv()
		if err != nil {
			es.errChan <- &GRPCStreamError{Err: err}
		} else {
			switch event := ev.Type.(type) {
			case *fabricPeer.DeliverResponse_Block:
				txFltr := util.TxValidationFlags(event.Block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
				for i, r := range event.Block.Data.Data {
					if txFltr.IsValid(i) {
						if ev, err := getCCEventFromEnvelope(r); err != nil {
							es.errChan <- &EnvelopeParsingError{Err: err}
						} else {
							if ev != nil {
								es.events <- ev
							}
						}
					} else {
						env, _ := utils.GetEnvelopeFromBlock(r)
						p, _ := utils.GetPayload(env)
						chHeader, _ := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
						es.errChan <- &InvalidTxError{TxId: api.ChaincodeTx(chHeader.TxId), Code: txFltr.Flag(i)}
					}
				}
			case *fabricPeer.DeliverResponse_FilteredBlock:
				es.errChan <- &UnknownEventTypeError{Type: `DeliverResponse_FilteredBlock`}
			default:
				es.errChan <- &UnknownEventTypeError{Type: fmt.Sprintf("%v", ev.Type)}
			}
		}
	}
}

func seekEnvelope(channelName string, startPos *orderer.SeekPosition, stopPos *orderer.SeekPosition, identity msp.SigningIdentity) (*common.Envelope, error) {
	creator, err := identity.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, `failed to get creator`)
	}

	channelHeader, err := proto.Marshal(&common.ChannelHeader{
		Type:      int32(common.HeaderType_DELIVER_SEEK_INFO),
		Version:   0,
		Timestamp: commonUtil.CreateUtcTimestamp(),
		ChannelId: channelName,
		Epoch:     0,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to marshal channel header`)
	}

	nonce, err := crypto.RandomBytes(24)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get nonce`)
	}

	signatureHeader, err := proto.Marshal(&common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to marshal signature header`)
	}

	seekData, err := proto.Marshal(&orderer.SeekInfo{
		Start:    startPos,
		Stop:     stopPos,
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to marshal seek info`)
	}

	payload, err := proto.Marshal(&common.Payload{
		Header: &common.Header{ChannelHeader: channelHeader, SignatureHeader: signatureHeader},
		Data:   seekData,
	})
	if err != nil {
		return nil, errors.Wrap(err, `failed to marshal payload`)
	}

	payloadSignature, err := identity.Sign(payload)
	if err != nil {
		return nil, errors.Wrap(err, `failed to sign payload`)
	}

	return &common.Envelope{Payload: payload, Signature: payloadSignature}, nil
}

func (es *eventSubscription) Close() error {
	return es.client.CloseSend()
}

type txSubscription struct {
	txId        api.ChaincodeTx
	client      fabricPeer.Deliver_DeliverClient
	events      chan api.TxEvent
	channelName string
	parentConn  *grpc.ClientConn
	identity    msp.SigningIdentity
}

func (ts *txSubscription) Result() (chan api.TxEvent, error) {
	var err error
	if ts.client, err = fabricPeer.NewDeliverClient(ts.parentConn).Deliver(context.Background()); err != nil {
		return nil, errors.Wrap(err, `failed to initialize deliver client`)
	}

	// Get only new events
	startPos, stopPos := api.SeekNewest()()

	if env, err := seekEnvelope(ts.channelName, startPos, stopPos, ts.identity); err != nil {
		return nil, errors.Wrap(err, `failed to get seek envelope`)
	} else {
		if err = ts.client.Send(env); err != nil {
			return nil, errors.Wrap(err, `failed to send seek envelope`)
		}
	}

	go ts.handleSubscription()

	return ts.events, nil
}

func (ts *txSubscription) handleSubscription() {
	for {
		ev, err := ts.client.Recv()
		if err != nil {
			log.Println(err)
		} else {
			switch event := ev.Type.(type) {
			case *fabricPeer.DeliverResponse_Block:
				txFltr := util.TxValidationFlags(event.Block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
				for i, r := range event.Block.Data.Data {
					env, err := utils.GetEnvelopeFromBlock(r)
					if err != nil {
						log.Println(`failed to get envelope:`, err)
						continue
					}

					p, err := utils.GetPayload(env)
					if err != nil {
						log.Println(`failed to get payload:`, err)
						continue
					}

					chHeader, err := utils.UnmarshalChannelHeader(p.Header.ChannelHeader)
					if err != nil {
						log.Println(`failed to get channel header:`, err)
						continue
					}

					if api.ChaincodeTx(chHeader.TxId) == ts.txId {
						if txFltr.IsValid(i) {
							ts.events <- api.TxEvent{
								TxId:    ts.txId,
								Success: true,
								Error:   nil,
							}
						} else {
							ts.events <- api.TxEvent{
								TxId:    ts.txId,
								Success: false,
								Error:   errors.Errorf("TxId validation code failed:%s", fabricPeer.TxValidationCode_name[int32(txFltr.Flag(i))]),
							}
						}
					}
				}
			case *fabricPeer.DeliverResponse_FilteredBlock:
				log.Println(`FilteredBlock not implemented`)
			}
		}
	}
}

func (ts *txSubscription) Close() error {
	return ts.client.CloseSend()
}
