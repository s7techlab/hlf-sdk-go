package deliver

import (
	"github.com/s7techlab/hlf-sdk-go/api"
	dtesting "github.com/s7techlab/hlf-sdk-go/peer/deliver/testing"
	"context"
	"fmt"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/peer"
	"testing"
	"time"
)

type mockSignIdentity struct {
	msp.SigningIdentity
	creator []byte
}

func (m *mockSignIdentity) Serialize() ([]byte, error) {
	return m.creator, nil
}

func (*mockSignIdentity) Sign(msg []byte) ([]byte, error) {
	return msg, nil
}

func TestDeliverImpl_SubscribeBlock(t *testing.T) {
	identity := &mockSignIdentity{
		creator: []byte(`ALOXA`),
	}

	deliverCli, err := dtesting.NewDeliverClient(`testdata/blocks`, true)
	if err != nil {
		t.Fatal(err)
	}

	deliverApiCli := New(deliverCli, identity)

	for _, tc := range []struct {
		channel     string
		seek        []api.EventCCSeekOption
		countBlocks int
	}{
		{
			channel: `payment-testbankmsp`,
			seek: []api.EventCCSeekOption{
				api.SeekOldest(),
			},
			countBlocks: 161,
		},
		{
			channel: `payment-testbankmsp`,
			seek: []api.EventCCSeekOption{
				api.SeekNewest(),
			},
			countBlocks: 1,
		},
		{
			channel: `payment-testbankmsp`,
			seek: []api.EventCCSeekOption{
				api.SeekRange(100, 110),
			},
			countBlocks: 11,
		},
	} {
		t.Run(fmt.Sprintf("%s-%d", tc.channel, tc.countBlocks), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			sub, err := deliverApiCli.SubscribeBlock(ctx, tc.channel, tc.seek...)
			if err != nil {
				t.Fatal(err)
			}

			var countBlocks int

			defer func() {
				if countBlocks != tc.countBlocks {
					t.Errorf("Unexpected block count %d != %d", tc.countBlocks, countBlocks)
				}
				if err = sub.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			for {
				select {
				case _, ok := <-sub.Blocks():
					if !ok {
						return
					}
					countBlocks++
				case err := <-sub.Errors():
					if err != nil {
						t.Fatal(err)
					} else {
						return
					}
				}
			}
		})
	}
}

func TestDeliverImpl_SubscribeCC(t *testing.T) {
	identity := &mockSignIdentity{
		creator: []byte(`ALOXA`),
	}

	deliverCli, err := dtesting.NewDeliverClient(`testdata/blocks`, true)
	if err != nil {
		t.Fatal(err)
	}

	deliverApiCli := New(deliverCli, identity)

	for _, tc := range []struct {
		channel     string
		seek        []api.EventCCSeekOption
		ccName      string
		countEvents int
	}{
		{
			channel: `payment-testbankmsp`,
			ccName:  `payment`,
			seek: []api.EventCCSeekOption{
				api.SeekOldest(),
			},
			countEvents: 5,
		},
	} {
		t.Run(fmt.Sprintf("%s/%s-%d", tc.channel, tc.ccName, tc.countEvents), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			sub, err := deliverApiCli.SubscribeCC(ctx, tc.channel, tc.ccName, tc.seek...)
			if err != nil {
				t.Fatal(err)
			}

			var countEvents int

			defer func() {
				if countEvents != tc.countEvents {
					t.Errorf("Unexpected events count %d != %d", tc.countEvents, countEvents)
				}
				if err = sub.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			for {
				select {
				case ev, ok := <-sub.Events():
					if !ok {
						return
					}
					t.Log("got event ", ev.EventName, ev.TxId)
					countEvents++
				case err := <-sub.Errors():
					if err != nil {
						t.Fatal(err)
					} else {
						return
					}
				}
			}
		})
	}
}

func TestDeliverImpl_SubscribeTx(t *testing.T) {
	identity := &mockSignIdentity{
		creator: []byte(`ALOXA`),
	}

	deliverCli, err := dtesting.NewDeliverClient(`testdata/blocks`, true)
	if err != nil {
		t.Fatal(err)
	}

	deliverApiCli := New(deliverCli, identity)

	for _, tc := range []struct {
		channel string
		seek    []api.EventCCSeekOption
		txid    api.ChaincodeTx
		code    peer.TxValidationCode
		err     error
	}{

		// TX from last blocks
		{
			channel: `payment-testbankmsp`,
			txid:    api.ChaincodeTx(`eb9f2fc22705f9d08e5a9df558e9486284cb3154297dc45d1c7bd727991f9fd6`),
			seek: []api.EventCCSeekOption{
				api.SeekNewest(),
			},
			code: peer.TxValidationCode_VALID,
		},
		{
			channel: `payment-testbankmsp`,
			txid:    api.ChaincodeTx(`9896fa40e2165311541156f1f6c6923de7c7913ac9d503c40bcac0da3f976338`),
			code:    peer.TxValidationCode_VALID,
		},
		{
			channel: `payment-testbankmsp`,
			txid:    api.ChaincodeTx(`38434cbd1f38b2b853f0e69160078f66839419c2d3f730b02dc649028153f99d`),
			code:    peer.TxValidationCode_VALID,
		},
		{
			channel: `payment-testbankmsp`,
			txid:    api.ChaincodeTx(`4192562e25a4c9d27e5f893732ada36b85bc25bfa38f89a0ec961c5f957e0342`),
			code:    peer.TxValidationCode_VALID,
		},
		{
			channel: `ticketing-s7testmsp`,
			txid:    api.ChaincodeTx(`never-got`),
			// WE GOR ERROR and code=-1
			code: -1,
		},
		// TX from oldes blocks
		{
			channel: `payment-testbankmsp`,
			txid:    api.ChaincodeTx(`542c3a0eefab36576316f4ddc803f2d1242f27b4f3d59d2b359037fe580f41f2`),
			seek: []api.EventCCSeekOption{
				api.SeekOldest(),
			},
			code: peer.TxValidationCode_VALID,
		},
		// TODO: add invalid blocks
	} {
		t.Run(fmt.Sprintf("%s-%s-%s", tc.channel, tc.code.String(), tc.txid), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			sub, err := deliverApiCli.SubscribeTx(ctx, tc.channel, tc.txid, tc.seek...)
			if err != nil {
				t.Fatal(err)
			}

			code, err := sub.Result()
			if code != tc.code {
				t.Errorf("Unexpected code %s != %s", tc.code.String(), code.String())
				if err != nil {
					t.Log(err.Error())
				}

			}

			if err = sub.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
