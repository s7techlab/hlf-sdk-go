package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

type (
	RegisterHandlerFromEndpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

	Def struct {
		name                        string
		swagger                     []byte
		Desc                        *grpc.ServiceDesc
		Service                     interface{}
		HandlerFromEndpointRegister RegisterHandlerFromEndpoint
	}
)

func NewDef(name string, swagger []byte, desc *grpc.ServiceDesc, service interface{}, registerHandler RegisterHandlerFromEndpoint) *Def {
	return &Def{
		name:                        name,
		swagger:                     swagger,
		Desc:                        desc,
		Service:                     service,
		HandlerFromEndpointRegister: registerHandler,
	}
}

func (s *Def) Name() string {
	return s.name
}

func (s *Def) Swagger() []byte {
	return s.swagger
}

func (s *Def) GRPCDesc() *grpc.ServiceDesc {
	return s.Desc
}

func (s *Def) Impl() interface{} {
	return s.Service
}

func (s *Def) GRPCGatewayRegister() func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	return s.HandlerFromEndpointRegister
}
