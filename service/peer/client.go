package peer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/client"
	"github.com/s7techlab/hlf-sdk-go/identity"
)

var (
	ErrAdminIdentityRequired = errors.New(`admin identity required`)
)

func Client(ctx context.Context, msp *identity.MSP, conn config.ConnectionConfig, endorseTimeout time.Duration, logger *zap.Logger) (*client.Client, error) {

	isFabricV2 := p.fabricVersion == hlfconfig.FabricV2

	peerPool := client.NewPeerPool(ctx, logger)

	if len(msp.Admins()) == 0 {
		return nil, fmt.Errorf(`no admins in provided msp: %w`, ErrAdminIdentityRequired)
	}

	admin := msp.Admins()[0]

	c, err := client.New(ctx,
		client.WithLogger(logger.With(zap.String("msp_id", admin.GetMSPIdentifier()))),
		// endorse only at this peer
		client.WithPeers(admin.GetMSPIdentifier(), []config.ConnectionConfig{
			{
				Host:    conn.Host,
				Timeout: endorseTimeout,
			},
		}),
	)

	//clientCore, err := client.New(admin,

	//	client.WithPeerPool(peerPool),
	//

	//	client.WithFabricV2(isFabricV2),
	//
	//	client.WithConfigRaw(sdkconfig.Config{
	//		Discovery: sdkconfig.DiscoveryConfig{
	//			Type: string(discovery.GossipServiceDiscoveryType),
	//			Connection: &sdkconfig.ConnectionConfig{
	//				Host: p.connection.URL,
	//			},
	//		},
	//	}),
	//)
	if err != nil {
		return nil, fmt.Errorf("create sdk client with identity msp_id=%s: %w", admin.GetMSPIdentifier(), err)
	}

	return c, nil

}
