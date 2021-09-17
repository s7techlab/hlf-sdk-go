package hlf

import (
	"context"
	"errors"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/stats"
)

// Easy way for plug-up grpc.StatsHandler and mixing needed annotations to span
func Wrap(statsHandler *ocgrpc.ClientHandler) stats.Handler {
	return &wrapper{
		oc: statsHandler,
	}
}

type wrapper struct {
	oc *ocgrpc.ClientHandler
}

func (w *wrapper) TagConn(ctx context.Context, connInfo *stats.ConnTagInfo) context.Context {
	if w.oc != nil {
		return w.oc.TagConn(ctx, connInfo)
	}

	return ctx
}

func (w *wrapper) HandleConn(ctx context.Context, cs stats.ConnStats) {
	if w.oc != nil {
		w.oc.HandleConn(ctx, cs)
	}
}

func (w *wrapper) TagRPC(ctx context.Context, rpcInfo *stats.RPCTagInfo) context.Context {
	if w.oc != nil {
		return w.oc.TagRPC(ctx, rpcInfo)
	}
	return ctx
}

func (w *wrapper) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	span := trace.FromContext(ctx)

	switch rs := rs.(type) {
	case *stats.InHeader:
		if rs.RemoteAddr != nil {
			span.AddAttributes(
				trace.StringAttribute(
					"InHeader.RemoteAddr",
					rs.RemoteAddr.String(),
				),
			)
		}
		if rs.LocalAddr != nil {
			span.AddAttributes(
				trace.StringAttribute(
					"InHeader.LocalAddr",
					rs.LocalAddr.String(),
				),
			)
		}

		span.AddAttributes(
			trace.StringAttribute(
				"InHeader.Compression",
				rs.Compression,
			),
		)
	case *stats.OutHeader:
		if rs.RemoteAddr != nil {
			span.AddAttributes(
				trace.StringAttribute(
					"OutHeader.RemoteAddr",
					rs.RemoteAddr.String(),
				),
			)
		}
		if rs.LocalAddr != nil {
			span.AddAttributes(
				trace.StringAttribute(
					"OutHeader.LocalAddr",
					rs.LocalAddr.String(),
				),
			)
		}

		span.AddAttributes(
			trace.StringAttribute(
				"OutHeader.Compression",
				rs.Compression,
			),
		)
	}
	// sometimes we get cancelled context if futher execution(asking peers etc.) isn't necessary
	// but request is fully valid and we dont want to see confusing errors in jaeger
	if errors.Is(ctx.Err(), context.Canceled) {
		return
	}

	if w.oc != nil {
		w.oc.HandleRPC(ctx, rs)
	}
}
