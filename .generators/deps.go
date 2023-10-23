//go:build tools
// +build tools

package generators

import (
	// proto/grpc
	_ "github.com/golang/protobuf/protoc-gen-go"

	// validation schema
	_ "github.com/envoyproxy/protoc-gen-validate"

	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"

	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
)
