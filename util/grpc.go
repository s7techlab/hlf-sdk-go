package util

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/opencensus/hlf"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

var (
	retryDefaultConfig = config.GRPCRetryConfig{
		Max:     10,
		Timeout: config.Duration{Duration: 10 * time.Second},
	}
)

const (
	maxRecvMsgSize = 100 * 1024 * 1024
	maxSendMsgSize = 100 * 1024 * 1024
)

func NewGRPCOptionsFromConfig(c config.ConnectionConfig, log *zap.Logger) ([]grpc.DialOption, error) {
	l := log.Named(`NewGRPCOptionsFromConfig`)

	// TODO: move to config or variable options
	grpcOptions := []grpc.DialOption{
		grpc.WithStatsHandler(hlf.Wrap(&ocgrpc.ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler:  trace.AlwaysSample(),
				SpanKind: trace.SpanKindClient,
			},
		})),
	}

	l.Debug(`Connection to host`, zap.String(`host`, c.Host))

	if c.Tls.Enabled {
		l.Debug(`Using TLS credentials`)
		var err error
		var tlsCfg tls.Config
		tlsCfg.InsecureSkipVerify = c.Tls.SkipVerify
		// if custom CA certificate is presented, use it
		if c.Tls.CACertPath != `` {
			log.Debug(`Using custom CA certificate`)
			caCert, err := ioutil.ReadFile(c.Tls.CACertPath)
			if err != nil {
				return nil, errors.Wrap(err, `failed to read CA certificate`)
			}
			certPool := x509.NewCertPool()
			if ok := certPool.AppendCertsFromPEM(caCert); !ok {
				return nil, errors.New(`failed to append CA certificate to chain`)
			}
			tlsCfg.RootCAs = certPool
		} else {
			log.Debug(`Using system CA certificates`)
			// otherwise we use system certificates
			if tlsCfg.RootCAs, err = x509.SystemCertPool(); err != nil {
				return nil, errors.Wrap(err, `failed to get system cert pool`)
			}
		}
		if c.Tls.CertPath != `` {
			// use mutual tls if certificate and pk is presented
			if c.Tls.KeyPath != `` {
				cert, err := tls.LoadX509KeyPair(c.Tls.CertPath, c.Tls.KeyPath)
				if err != nil {
					return nil, errors.Wrap(err, `failed to load client certificate`)
				}
				tlsCfg.Certificates = append(tlsCfg.Certificates, cert)
			} else {
				// otherwise just add cert to pool
				//cert, err := ioutil.ReadFile(c.Tls.CertPath)
				//if err != nil {
				//	return nil, errors.Wrap(err, `failed to read server certificate`)
				//}
				//if ok := tlsCfg.RootCAs.AppendCertsFromPEM(cert); !ok {
				//	return nil, errors.New(`failed to add server certificate to chain`)
				//}
				//log.Debug(`Using server certificate`)
			}
		}

		cred := credentials.NewTLS(&tlsCfg)
		l.Debug(`Read TLS credentials`, zap.Reflect(`cred`, cred.Info()))
		grpcOptions = append(grpcOptions, grpc.WithTransportCredentials(cred))
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
			Time:                time.Duration(c.GRPC.KeepAlive.Time) * time.Second,
			Timeout:             time.Duration(c.GRPC.KeepAlive.Timeout) * time.Second,
			PermitWithoutStream: true,
		}))
	} else {
		grpcOptions = append(grpcOptions, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
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

	grpcOptions = append(grpcOptions, grpc.WithReturnConnectionError())

	return grpcOptions, nil
}

func NewGRPCConnectionFromConfigs(ctx context.Context, log *zap.Logger, conf ...config.ConnectionConfig) (*grpc.ClientConn, error) {
	// use options from first config
	opts, err := NewGRPCOptionsFromConfig(conf[0], log)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get GRPC options`)
	}

	addr := make([]resolver.Address, len(conf))
	for i, cc := range conf {
		addr[i] = resolver.Address{Addr: cc.Host}
	}

	r, _ := manual.GenerateAndRegisterManualResolver()
	r.InitialState(resolver.State{Addresses: addr})

	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:///%s", r.Scheme(), `orderers`), opts...)
	if err != nil {
		return nil, errors.Wrap(err, `failed to initialize GRPC connection`)
	}

	return conn, nil
}
