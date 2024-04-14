package chaincode_test

// tests temporary disabled

//import (
//	"context"
//	"fmt"
//	"strings"
//	"testing"
//
//	"github.com/golang/protobuf/proto"
//	"github.com/hyperledger/fabric-protos-go/common"
//	"github.com/hyperledger/fabric-protos-go/orderer"
//	"github.com/hyperledger/fabric-protos-go/peer"
//	"github.com/hyperledger/fabric/msp"
//	"github.com/hyperledger/fabric/protoutil"
//	"github.com/pkg/errors"
//	"google.golang.org/grpc"
//
//	"github.com/s7techlab/hlf-sdk-go/api"
//	"github.com/s7techlab/hlf-sdk-go/client"
//	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
//	"github.com/s7techlab/hlf-sdk-go/client/chaincode/txwaiter"
//	"github.com/s7techlab/hlf-sdk-go/identity"
//	"github.com/s7techlab/hlf-sdk-go/logger"
//	"github.com/s7techlab/hlf-sdk-go/peer/pool"
//)
//
//var (
//	_ api.Peer    = &mockPeer{}
//	_ api.Orderer = &mockOrderer{}
//)
//
//// simple mock orderer
//type mockOrderer struct {
//}
//
//func (m *mockOrderer) Broadcast(ctx context.Context, envelope *common.Envelope) (*orderer.BroadcastResponse, error) {
//	return nil, nil
//}
//func (m *mockOrderer) Deliver(ctx context.Context, envelope *common.Envelope) (*common.Block, error) {
//	return nil, nil
//}
//
//// simple mock deliver
//func newMockDeliverClient(channelConfig map[string]deliverChannelRouter) *mockDeliverClient {
//	return &mockDeliverClient{
//		txResultCount: make(map[string]int, 0),
//		channelConfig: channelConfig,
//	}
//}
//
//type deliverChannelRouter struct {
//	errForMakeSubscribeTx error
//	txCode                peer.TxValidationCode
//}
//
//type mockDeliverClient struct {
//	txResultCount map[string]int
//	channelConfig map[string]deliverChannelRouter
//}
//
//func (m *mockDeliverClient) SubscribeCC(ctx context.Context, channelName string, ccName string, seekOpt ...api.EventCCSeekOption) (api.EventCCSubscription, error) {
//	return nil, nil
//}
//func (m *mockDeliverClient) SubscribeTx(ctx context.Context, channelName string, tx api.ChaincodeTx, seekOpt ...api.EventCCSeekOption) (api.TxSubscription, error) {
//	cfg := m.channelConfig[channelName]
//	return &mockTxSubscription{
//		channelName: channelName,
//		tx:          tx,
//		txCode:      cfg.txCode,
//		deliver:     m,
//	}, cfg.errForMakeSubscribeTx
//}
//
//type mockTxSubscription struct {
//	deliver     *mockDeliverClient
//	channelName string
//	tx          api.ChaincodeTx
//	txCode      peer.TxValidationCode
//}
//
//func (t *mockTxSubscription) Inc() {
//	key := t.channelName + `/` + string(t.tx)
//	t.deliver.txResultCount[key]++
//}
//
//func (t *mockTxSubscription) Result() (peer.TxValidationCode, error) {
//	t.Inc()
//	if t.txCode != peer.TxValidationCode_VALID {
//		err := fmt.Errorf("TxId validation code failed: %s", peer.TxValidationCode_name[int32(t.txCode)])
//		println(err.Error())
//		return t.txCode, err
//	}
//	return t.txCode, nil
//}
//
//func (t *mockTxSubscription) Close() error {
//	return nil
//}
//
//func (m *mockDeliverClient) SubscribeBlock(ctx context.Context, channelName string, seekOpt ...api.EventCCSeekOption) (api.BlockSubscription, error) {
//	return nil, nil
//}
//
//// simple mock peer
//type mockPeer struct {
//	deliver      *mockDeliverClient
//	endorser     msp.SigningIdentity
//	checkEndorse map[string]int
//}
//
//func (p *mockPeer) Query(ctx context.Context, chanName string, ccName string, args [][]byte, identity msp.SigningIdentity, transient map[string][]byte) (*peer.Response, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//// Endorse mock echo answer from peer
//func (p *mockPeer) Endorse(ctx context.Context, proposal *peer.SignedProposal) (*peer.ProposalResponse, error) {
//	prop := new(peer.Proposal)
//
//	if err := proto.Unmarshal(proposal.ProposalBytes, prop); err != nil {
//		return nil, errors.Wrap(err, `failed to unmarshal ProposalBytes`)
//	}
//	header, err := protoutil.UnmarshalHeader(prop.Header)
//	if err != nil {
//		return nil, errors.Wrap(err, `failed to unmarshal Header`)
//	}
//	chHeader, err := protoutil.UnmarshalChannelHeader(header.ChannelHeader)
//	if err != nil {
//		return nil, errors.Wrap(err, `failed to unmarshal`)
//	}
//
//	p.checkEndorse[chHeader.ChannelId+`/`+chHeader.TxId]++
//
//	peerResp := &peer.Response{
//		Status:  200,
//		Payload: []byte(`{"message": "OK"}`),
//	}
//
//	result := []byte(``)
//	event := []byte(nil)
//	ccId := &peer.ChaincodeID{
//		Name:    `my-chaincode`,
//		Version: `0.1`,
//	}
//
//	return protoutil.CreateProposalResponse(
//		prop.Header,
//		prop.Payload,
//		peerResp,
//		result,
//		event,
//		ccId,
//		p.endorser,
//	)
//}
//
//func (p *mockPeer) DeliverClient(id msp.SigningIdentity) (api.DeliverClient, error) {
//	return p.deliver, nil
//}
//
//// URI returns url used for grpc connection
//func (p *mockPeer) URI() string {
//	return "localhost:7051"
//}
//
//// Conn returns instance of grpc connection
//func (p *mockPeer) Conn() *grpc.ClientConn {
//	return nil
//}
//
//// Close terminates peer connection
//func (p *mockPeer) Close() error {
//	return nil
//}
//
//func defaultAlivePeer(_ context.Context, _ api.Peer, alive chan bool) {
//	alive <- true
//	return
//}
//
//func TestInvokeBuilder_Do(t *testing.T) {
//	//get identity
//	org1mspID, err := identity.SignerFromMSPPath(`org1msp`, `./testdata/msp`)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	org2mspID, err := identity.SignerFromMSPPath(`org2msp`, `./testdata/msp`)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	org3mspID, err := identity.SignerFromMSPPath(`org3msp`, `./testdata/msp`)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	cryptoSuite := client.DefaultCryptoSuite()
//
//	channelConfigPeer1And2 := map[string]deliverChannelRouter{
//		"success-network": {
//			errForMakeSubscribeTx: nil,
//			txCode:                peer.TxValidationCode_VALID,
//		},
//		"fail-network": {
//			errForMakeSubscribeTx: errors.New(`BOOM`),
//			txCode:                peer.TxValidationCode_VALID,
//		},
//		"fail-invalid-org3-network": {
//			errForMakeSubscribeTx: nil,
//			txCode:                peer.TxValidationCode_VALID,
//		},
//		"fail-mvcc-network": {
//			errForMakeSubscribeTx: nil,
//			txCode:                peer.TxValidationCode_MVCC_READ_CONFLICT,
//		},
//	}
//	channelConfigPeer3 := map[string]deliverChannelRouter{
//		"success-network": {
//			errForMakeSubscribeTx: nil,
//			txCode:                peer.TxValidationCode_VALID,
//		},
//		"fail-network": {
//			errForMakeSubscribeTx: errors.New(`BOOM`),
//			txCode:                peer.TxValidationCode_VALID,
//		},
//		"fail-invalid-org3-network": {
//			errForMakeSubscribeTx: nil,
//			txCode:                peer.TxValidationCode_BAD_PAYLOAD,
//		},
//		"fail-mvcc-network": {
//			errForMakeSubscribeTx: nil,
//			txCode:                peer.TxValidationCode_MVCC_READ_CONFLICT,
//		},
//	}
//
//	var (
//		peerOrg1 = &mockPeer{
//			deliver:      newMockDeliverClient(channelConfigPeer1And2),
//			endorser:     org1mspID.GetSigningIdentity(cryptoSuite),
//			checkEndorse: make(map[string]int),
//		}
//
//		peerOrg2 = &mockPeer{
//			deliver:      newMockDeliverClient(channelConfigPeer1And2),
//			endorser:     org2mspID.GetSigningIdentity(cryptoSuite),
//			checkEndorse: make(map[string]int),
//		}
//
//		peerOrg3 = &mockPeer{
//			deliver:      newMockDeliverClient(channelConfigPeer3),
//			endorser:     org3mspID.GetSigningIdentity(cryptoSuite),
//			checkEndorse: make(map[string]int),
//		}
//
//		peerResolver = map[string]*mockPeer{
//			"org1msp": peerOrg1,
//			"org2msp": peerOrg2,
//			"org3msp": peerOrg3,
//		}
//	)
//
//	peerPool := pool.New(context.Background(), logger.DefaultLogger)
//	_ = peerPool.Add(`org1msp`, peerOrg1, defaultAlivePeer)
//	_ = peerPool.Add(`org2msp`, peerOrg2, defaultAlivePeer)
//	_ = peerPool.Add(`org3msp`, peerOrg3, defaultAlivePeer)
//
//	core, err := client.NewCore(
//		org1mspID,
//		client.WithOrderer(&mockOrderer{}),
//		client.WithPeerPool(peerPool),
//		client.WithConfigYaml(`./testdata/config.yaml`),
//	)
//
//
//	var checkEndorsingCount = func(channelName, txId string, orgs ...string) error {
//		var strErrs []string
//		key := channelName + `/` + txId
//		for _, org := range orgs {
//			if v := peerResolver[org].checkEndorse[key]; v != 1 {
//				strErrs = append(strErrs, fmt.Sprintf(
//					"expected endorse was called on peer %s/%s/%d != 1",
//					org,
//					key,
//					v,
//				))
//			}
//		}
//
//		if len(strErrs) != 0 {
//			return fmt.Errorf("%s", strings.Join(strErrs, "\n"))
//		}
//
//		return nil
//	}
//
//	var checkTxWaiterCount = func(channelName, txId string, orgs ...string) error {
//		var strErrs []string
//		key := channelName + `/` + txId
//		for _, org := range orgs {
//			if v := peerResolver[org].deliver.txResultCount[key]; v != 1 {
//				strErrs = append(strErrs, fmt.Sprintf(
//					"expected wait of tx was called on peer %s/%s/%d != 1",
//					org,
//					key,
//					v,
//				))
//			}
//		}
//
//		if len(strErrs) != 0 {
//			return fmt.Errorf("%s", strings.Join(strErrs, "\n"))
//		}
//
//		return nil
//	}
//
//	for _, tc := range []struct {
//		name                   string
//		opts                   []api.DoOption
//		channel                string
//		chaincode              string
//		checkEndorseCalled     []string
//		checkDeliverByTxCalled []string
//		expErr                 error
//	}{
//		{
//			name:                   `success with self peer`,
//			channel:                `success-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{`org1msp`},
//		},
//		{
//			name:                   `success with self peer directly`,
//			channel:                `success-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.Self)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{`org1msp`},
//		},
//		{
//			name:                   `fail self peer on make deliver`,
//			channel:                `fail-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.Self)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{},
//			expErr:                 errors.New(`org1msp: failed to subscribe on tx event: BOOM`),
//		},
//		{
//			name:                   `fail validation self peer by mvcc`,
//			channel:                `fail-mvcc-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.Self)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{`org1msp`},
//			expErr:                 errors.New(`TxId validation code failed: MVCC_READ_CONFLICT`),
//		},
//		{
//			name:                   `success with all peer`,
//			channel:                `success-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.All)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{`org1msp`, `org2msp`, `org3msp`},
//		},
//		{
//			name:                   `fail tx validation on all peer`,
//			channel:                `fail-mvcc-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.All)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{`org1msp`, `org2msp`, `org3msp`},
//			expErr:                 errors.New("next errors occurred:\nTxId validation code failed: MVCC_READ_CONFLICT\nTxId validation code failed: MVCC_READ_CONFLICT\nTxId validation code failed: MVCC_READ_CONFLICT\n"),
//		},
//		{
//			name:                   `fail tx validation on org3 for wait from all peer`,
//			channel:                `fail-invalid-org3-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.All)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{`org1msp`, `org2msp`, `org3msp`},
//			expErr:                 errors.New("next errors occurred:\nTxId validation code failed: BAD_PAYLOAD\n"),
//		},
//		{
//			name:                   `fail all peer on make deliver`,
//			channel:                `fail-network`,
//			chaincode:              `my-chaincode`,
//			opts:                   []api.DoOption{chaincode.WithTxWaiter(txwaiter.All)},
//			checkEndorseCalled:     []string{`org1msp`, `org2msp`, `org3msp`},
//			checkDeliverByTxCalled: []string{},
//			expErr:                 errors.New("next errors occurred:\nfailed to subscribe on tx event: BOOM\nfailed to subscribe on tx event: BOOM\nfailed to subscribe on tx event: BOOM\n"),
//		},
//	} {
//		t.Run(tc.name, func(tt *testing.T) {
//			_, txId, err := core.Invoke(
//				context.Background(),
//				org1mspID.GetSigningIdentity(cryptoSuite),
//				tc.channel,
//				tc.chaincode,
//				`call`,
//				nil,
//				nil,
//				tc.opts...,
//			)
//			//_, txId, err := NewInvokeBuilder(mockCore, `call`).Do(ctx, tc.opts...)
//			if fmt.Sprint(tc.expErr) != fmt.Sprint(err) {
//				tt.Errorf("Unexpected error:\n %s \n!=\n %s", tc.expErr, err)
//			}
//			if len(tc.checkEndorseCalled) != 0 {
//				if err = checkEndorsingCount(tc.channel, string(txId), tc.checkEndorseCalled...); err != nil {
//					t.Errorf("checkEndorseCalled: Unexpected error: %s", err)
//				}
//			}
//			if len(tc.checkDeliverByTxCalled) != 0 {
//				if err = checkTxWaiterCount(tc.channel, string(txId), tc.checkDeliverByTxCalled...); err != nil {
//					t.Errorf("checkDeliverByTxCalled: Unexpected error: %s", err)
//				}
//			}
//		})
//	}
//}
