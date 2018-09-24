package peer

import (
	"time"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func NewGRPCFromConfig(c config.PeerConfig, log *zap.Logger) (*grpc.ClientConn, error) {
	l := log.Named(`NewGRPCFromConfig`)
	var grpcOptions []grpc.DialOption

	if c.Tls.Enabled {
		l.Debug(`Using TLS credentials`)
		if ts, err := credentials.NewClientTLSFromFile(c.Tls.CertPath, ``); err != nil {
			l.Error(`Failed to read TLS credentials`, zap.Error(err))
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			l.Debug(`Read TLS credentials`, zap.Reflect(`cred`, ts.Info()))
			grpcOptions = append(grpcOptions, grpc.WithTransportCredentials(ts))
		}
	} else {
		l.Debug(`TLS is not used`)
		grpcOptions = append(grpcOptions, grpc.WithInsecure())
	}

	// Set KeepAlive parameters if presented
	if c.GRPC.KeepAlive != nil {
		l.Debug(`Using KeepAlive params`,
			zap.Duration(`time`, time.Duration(c.GRPC.KeepAlive.Time)*time.Second),
			zap.Duration(`timeout`, time.Duration(c.GRPC.KeepAlive.Timeout)*time.Second),
		)
		grpcOptions = append(grpcOptions, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Duration(c.GRPC.KeepAlive.Time) * time.Second,
			Timeout: time.Duration(c.GRPC.KeepAlive.Timeout) * time.Second,
		}))
	}

	grpcOptions = append(grpcOptions, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	return grpc.Dial(c.Host, grpcOptions...)
}
