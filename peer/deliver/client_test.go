package deliver

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/crypto"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	id     msp.SigningIdentity
	suite  api.CryptoSuite
	logger *zap.Logger

	conn *grpc.ClientConn
	cli  api.DeliverClient
)

func TestNewFromGRPC(t *testing.T) {
	cli = NewFromGRPC(context.Background(), conn, id, logger)
	assert.NotNil(t, cli)
}

func TestDeliverClient_SubscribeBlock(t *testing.T) {
	sub, err := cli.SubscribeBlock(context.Background(), os.Getenv(`CHANNEL`))
	assert.NotNil(t, sub)
	assert.NoError(t, err)

	assert.NotNil(t, sub.Errors())
	assert.NotNil(t, sub.Blocks())
	assert.NoError(t, sub.Close())
}

func TestDeliverClient_SubscribeCC(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sub, err := cli.SubscribeCC(ctx, os.Getenv(`CHANNEL`), os.Getenv(`CHAINCODE`))
	assert.NotNil(t, sub)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

forLoop:
	for {
		select {
		case err := <-sub.Errors():
			assert.Equal(t, err, context.Canceled)
			break forLoop
		case <-sub.Events():
		}
	}

	assert.NoError(t, sub.Close())
}

func TestDeliverClient_SubscribeTx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sub, err := cli.SubscribeTx(ctx, os.Getenv(`CHANNEL`), `someTxId`)
	assert.NotNil(t, sub)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	res, err := sub.Result()
	assert.Equal(t, res, peer.TxValidationCode(-1))
	assert.Error(t, err)
	assert.Equal(t, err, context.Canceled)

	assert.NoError(t, sub.Close())
}

func TestDeliverClient_SubscribeBlock2(t *testing.T) {
	sub, err := cli.SubscribeBlock(context.Background(), os.Getenv(`CHANNEL`))
	assert.NotNil(t, sub)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Second)
		conn.Close()
	}()

forLoop:
	for {
		select {
		case <-sub.Blocks():
		case err := <-sub.Errors():
			assert.IsType(t, &api.GRPCStreamError{}, err)
			break forLoop
		}
	}

	assert.NoError(t, sub.Close())
}

func TestDeliverClient_Close(t *testing.T) {
	assert.Error(t, cli.Close())
}

func init() {
	var err error
	suite, err = crypto.GetSuite(ecdsa.Module, ecdsa.DefaultOpts)
	if err != nil {
		log.Fatalln(err)
	}

	idd, err := identity.NewMSPIdentityFromPath(os.Getenv(`MSP_ID`), os.Getenv(`MSP_PATH`))
	if err != nil {
		log.Fatalln(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	conn, err = grpc.DialContext(ctx, os.Getenv(`PEER_HOST`), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalln(err)
	}

	id = idd.GetSigningIdentity(suite)

	logger, _ = zap.NewDevelopment()
}
