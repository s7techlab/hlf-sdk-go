package util

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	retryDefaultConfig = config.GRPCRetryConfig{
		Max:     10,
		Timeout: config.Duration{Duration: time.Second},
	}
)

const (
	maxRecvMsgSize = 1000 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

func NewGRPCOptionsFromConfig(c config.ConnectionConfig, log *zap.Logger) ([]grpc.DialOption, error) {
	l := log.Named(`NewGRPCOptionsFromConfig`)
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

	var retryConfig *config.GRPCRetryConfig
	if c.GRPC.Retry != nil {
		l.Debug(`Using presented GRPC retry config`, zap.Reflect(`config`, *c.GRPC.Retry))
		retryConfig = c.GRPC.Retry
	} else {
		l.Debug(`Using default GRPC retry config`, zap.Reflect(`config`, retryDefaultConfig))
		retryConfig = &retryDefaultConfig
	}

	grpcOptions = append(grpcOptions,
		grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithMax(retryConfig.Max),
				grpc_retry.WithBackoff(grpc_retry.BackoffLinear(retryConfig.Timeout.Duration)),
			),
		),
	)

	grpcOptions = append(grpcOptions, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxSendMsgSize),
	))

	return grpcOptions, nil
}
