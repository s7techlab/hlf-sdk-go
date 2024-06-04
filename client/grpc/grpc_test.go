package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	testpb "google.golang.org/grpc/interop/grpc_testing"
)

var (
	tlsListener       net.Listener
	tlsMutualListener net.Listener
	nonTlsListener    net.Listener

	log, _            = zap.NewDevelopment()
	_, filename, _, _ = runtime.Caller(0)
	curDir            = filepath.Dir(filename)

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
)

type testServer struct {
	testpb.TestServiceServer
}

func (s *testServer) EmptyCall(_ context.Context, _ *testpb.Empty) (*testpb.Empty, error) {
	return &testpb.Empty{}, nil
}

func (s *testServer) FullDuplexCall(_ testpb.TestService_FullDuplexCallServer) error {
	return nil
}

func TestNewGRPCOptionsFromConfig(t *testing.T) {

	defer func() {
		_ = nonTlsListener.Close()
		_ = tlsListener.Close()
		_ = tlsMutualListener.Close()
	}()
	// Testing non-tls connection
	nonTlsConnConfig := config.ConnectionConfig{
		Host: nonTlsListener.Addr().String(),
		Tls: config.TlsConfig{
			Enabled: false,
		},
		GRPC:    config.GRPCConfig{},
		Timeout: config.Duration{},
	}
	opts, err := OptionsFromConfig(nonTlsConnConfig, log)
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	conn, err := grpc.DialContext(ctx, nonTlsConnConfig.Host, opts.Dial...)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	cli := testpb.NewTestServiceClient(conn)
	out, err := cli.EmptyCall(context.Background(), &testpb.Empty{})
	assert.NotNil(t, out)
	assert.NoError(t, err)
	// Testing usual tls connection
	tlsConnConfig := config.ConnectionConfig{
		Host: getLocalAddress(tlsListener),
		Tls: config.TlsConfig{
			Enabled:    true,
			CACertPath: path.Join(curDir, `testdata`, `tls`, `ca`, `ca.pem`),
			CertPath:   path.Join(curDir, `testdata`, `tls`, `server`, `cert.pem`),
		},
	}
	opts, err = OptionsFromConfig(tlsConnConfig, log)
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	conn, err = grpc.DialContext(ctx, tlsConnConfig.Host, opts.Dial...)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	cli = testpb.NewTestServiceClient(conn)
	out, err = cli.EmptyCall(context.Background(), &testpb.Empty{})
	assert.NotNil(t, out)
	assert.NoError(t, err)
	// Testing mutual tls connection
	mutualTlsConnConfig := config.ConnectionConfig{
		Host: getLocalAddress(tlsMutualListener),
		Tls: config.TlsConfig{
			Enabled:    true,
			CACertPath: path.Join(curDir, `testdata`, `tls`, `ca`, `ca.pem`),
			CertPath:   path.Join(curDir, `testdata`, `tls`, `client`, `cert.pem`),
			KeyPath:    path.Join(curDir, `testdata`, `tls`, `client`, `cert-key.pem`),
		},
	}
	opts, err = OptionsFromConfig(mutualTlsConnConfig, log)
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	conn, err = grpc.DialContext(ctx, mutualTlsConnConfig.Host, opts.Dial...)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	cli = testpb.NewTestServiceClient(conn)
	out, err = cli.EmptyCall(context.Background(), &testpb.Empty{})
	assert.NotNil(t, out)
	assert.NoError(t, err)
}

func init() {
	var err error

	nonTlsListener, err = net.Listen(`tcp4`, `:`)
	if err != nil {
		panic(err)
	}
	nonTlsGrpcSrv := grpc.NewServer()
	testpb.RegisterTestServiceServer(nonTlsGrpcSrv, &testServer{})
	go func() {
		_ = nonTlsGrpcSrv.Serve(nonTlsListener)
	}()

	tlsListener, err = net.Listen(`tcp4`, `:`)
	if err != nil {
		panic(err)
	}

	creds, err := credentials.NewServerTLSFromFile(path.Join(curDir, `testdata`, `tls`, `server`, `cert.pem`), path.Join(curDir, `testdata`, `tls`, `server`, `cert-key.pem`))
	if err != nil {
		panic(err)
	}

	tlsGrpcSrv := grpc.NewServer(grpc.Creds(creds))
	testpb.RegisterTestServiceServer(tlsGrpcSrv, &testServer{})
	go func() {
		_ = tlsGrpcSrv.Serve(tlsListener)
	}()

	tlsMutualListener, err = net.Listen(`tcp4`, `:`)
	if err != nil {
		panic(err)
	}
	certificate, err := tls.LoadX509KeyPair(path.Join(curDir, `testdata`, `tls`, `server`, `cert.pem`), path.Join(curDir, `testdata`, `tls`, `server`, `cert-key.pem`))
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(path.Join(curDir, `testdata`, `tls`, `ca`, `ca.pem`))
	if err != nil {
		panic(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic("failed to append client certs")
	}

	creds = credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	tlsMutualGrpcSrv := grpc.NewServer(grpc.Creds(creds))
	testpb.RegisterTestServiceServer(tlsMutualGrpcSrv, &testServer{})
	go func() {
		_ = tlsMutualGrpcSrv.Serve(tlsMutualListener)
	}()

}

func getLocalAddress(lis net.Listener) string {
	addr := lis.Addr().String()
	return `localhost:` + strings.Split(addr, `:`)[1]
}
