package gateway

import (
	zfg "github.com/chaindead/zerocfg"
	"github.com/hughbliss/my_toolkit/telemetry/tracer/trace_propagator"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	connectionsGroup            = zfg.NewGroup("connection")
	ConnectionStringSomeService = zfg.Str("some_service", "0.0.0.0:11000", "CONNECTION_SOMESERVICE", zfg.Group(connectionsGroup))
	ConnectionStringAuthService = zfg.Str("auth_service", "0.0.0.0:12000", "CONNECTION_AUTHSERVICE", zfg.Group(connectionsGroup))

	// example declaring connection strings config
	//connectionStringYetAnotherService = zfg.Str("yet_another_service", "0.0.0.0:11000", "CONNECTION_YETANOTHERSERVICE", zfg.Group(connectionsGroup))

	DefaultGRPCOptions = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			otelgrpc.WithMeterProvider(otel.GetMeterProvider()),
		)),
		grpc.WithStatsHandler(trace_propagator.ClientTracePropagator()),
	}
)
