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
		if ts, err := credentials.NewClientTLSFromFile(c.Tls.CertPath, ``); err != nil {
			return nil, errors.Wrap(err, `failed to read tls credentials`)
		} else {
			grpcOptions = append(grpcOptions, grpc.WithTransportCredentials(ts))
		}
	} else {
		grpcOptions = append(grpcOptions, grpc.WithInsecure())
	}

	// Set KeepAlive parameters if presented
	if c.GRPC.KeepAlive != nil {
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
