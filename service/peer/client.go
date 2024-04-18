package peer

import (
	"errors"
)

var (
	ErrAdminIdentityRequired = errors.New(`admin identity required`)
)

//func InfoClient(ctx context.Context, conn config.ConnectionConfig, msp *identity.MSP, endorseTimeout time.Duration, logger *zap.Logger) (*client.Client, error) {
//	//
//	//isFabricV2 := p.fabricVersion == hlfconfig.FabricV2
//	//
//	//peerPool := client.NewPeerPool(ctx, logger)
//
//	//if len(msp.Admins()) == 0 {
//	//	return nil, fmt.Errorf(`no admins in provided msp: %w`, ErrAdminIdentityRequired)
//	//}
//
//	//admin := msp.Admins()[0]
//
//	//signer := msp.Signer()
//
//	//c, err := client.New(ctx,
//	//	client.WithLogger(logger.With(zap.String("msp_id", signer.GetMSPIdentifier()))),
//	//	// endorse only at this peer
//	//	client.WithPeers(signer.GetMSPIdentifier(), []config.ConnectionConfig{
//	//		{
//	//			Host:    conn.Host,
//	//			Timeout: endorseTimeout,
//	//		},
//	//	}),
//	//)
//
//	//return conn config.ConnectionConfig
//
//	//clientCore, err := client.New(admin,
//
//	//	client.WithPeerPool(peerPool),
//	//
//
//	//	client.WithFabricV2(isFabricV2),
//	//
//	//	client.WithConfigRaw(sdkconfig.Config{
//	//		Discovery: sdkconfig.DiscoveryConfig{
//	//			Type: string(discovery.GossipServiceDiscoveryType),
//	//			Connection: &sdkconfig.ConnectionConfig{
//	//				Host: p.connection.URL,
//	//			},
//	//		},
//	//	}),
//	//)
//	//if err != nil {
//	//	return nil, fmt.Errorf("create sdk client with identity msp_id=%s: %w", signer.GetMSPIdentifier(), err)
//	//}
//	//
//	//return c, nil
//
//}
