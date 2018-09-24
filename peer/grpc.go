package peer

import (
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func NewGRPCFromConfig(c config.PeerConfig, log *zap.Logger) (*grpc.ClientConn, error) {
	var (
		err         error
		grpcOptions []grpc.DialOption
	)

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

	grpcOptions = append(grpcOptions, grpc.WithBlock(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	if len(c.Hosts) == 0 {
		return nil, errors.Wrap(err, `no peer endpoints`)
	} else {
		grpcOptions = append(grpcOptions, grpc.WithBalancerName(roundrobin.Name))
	}

	return grpc.Dial(strings.Join(c.Hosts, `,`), grpcOptions...)
}
