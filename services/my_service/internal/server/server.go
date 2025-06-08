package server

import (
	"github.com/hughbliss/my_toolkit/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
)

func Init() *grpc.Server {
	return grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Timeout:           15 * time.Second,
			MaxConnectionAge:  5 * time.Minute,
			Time:              15 * time.Minute,
		}),

		grpc.StatsHandler(tracer.ServerTraceProvider()),
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tracer.Provider))),
	)
}
