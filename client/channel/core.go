package channel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode"
	"github.com/s7techlab/hlf-sdk-go/peer"
)

type Core struct {
	mspId        string
	chanName     string
	peerPool     api.PeerPool
	orderer      api.Orderer
	chaincodes   map[string]*chaincode.Core
	chaincodesMx sync.Mutex
	dp           api.DiscoveryProvider
	identity     msp.SigningIdentity
	fabricV2     bool
	log          *zap.Logger
}

var _ api.Channel = (*Core)(nil)

// Chaincode - returns interface with actions over chaincode
// ctx is necessary for service discovery
func (c *Core) Chaincode(serviceDiscCtx context.Context, ccName string) (api.Chaincode, error) {
	c.chaincodesMx.Lock()
	defer c.chaincodesMx.Unlock()

	cc, ok := c.chaincodes[ccName]
	if ok {
		return cc, nil
	}

	if c.chanName == `` {
		cc = chaincode.NewCore(c.mspId, ccName, c.chanName, []string{c.mspId}, c.peerPool, c.orderer, c.identity)
		c.chaincodes[ccName] = cc

		return cc, nil
	}

	cd, err := c.dp.Chaincode(serviceDiscCtx, c.chanName, ccName)
	if err != nil {
		return nil, fmt.Errorf("chaincode discovery: %w", err)
	}

	var endorserMSPs []string
	endorsers := cd.Endorsers()
	errGr, _ := errgroup.WithContext(serviceDiscCtx)

	for i := range endorsers {
		endorserMSPs = append(endorserMSPs, endorsers[i].MspID)

		for j := range endorsers[i].HostAddresses {
			hostAddr := endorsers[i].HostAddresses[j]
			// we can get empty address in local discovery and peers must be already in pool
			if hostAddr.Address == "" {
				continue
			}
			mspID := endorsers[i].MspID
			grpcCfg := config.ConnectionConfig{
				Host: hostAddr.Address,
				Tls:  hostAddr.TLSSettings,
			}
			l := c.log

			errGr.Go(func() error {
				var p api.Peer
				p, err = peer.New(serviceDiscCtx, grpcCfg, c.identity, l)
				if err != nil {
					return fmt.Errorf("initialize endorsers for MSP: %s: %w", mspID, err)
				}
				if err = c.peerPool.Add(mspID, p, api.StrategyGRPC(5*time.Second)); err != nil {
					return fmt.Errorf("add endorser peer to pool: %s:%w", mspID, err)
				}
				return nil
			})
		}
	}

	if err = errGr.Wait(); err != nil {
		return nil, err
	}

	cc = chaincode.NewCore(c.mspId, ccName, c.chanName, endorserMSPs, c.peerPool, c.orderer, c.identity)
	c.chaincodes[ccName] = cc

	return cc, nil
}

func NewCore(
	mspId, chanName string,
	peerPool api.PeerPool,
	orderer api.Orderer,
	dp api.DiscoveryProvider,
	identity msp.SigningIdentity,
	fabricV2 bool,
	log *zap.Logger,
) api.Channel {
	return &Core{
		mspId:      mspId,
		chanName:   chanName,
		peerPool:   peerPool,
		orderer:    orderer,
		chaincodes: make(map[string]*chaincode.Core),
		dp:         dp,
		identity:   identity,
		fabricV2:   fabricV2,
		log:        log,
	}
}
