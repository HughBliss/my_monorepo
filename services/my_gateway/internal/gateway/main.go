package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	perserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/permissions/v1"
	someservicev1 "github.com/hughbliss/my_protobuf/go/pkg/gen/someservice/v1"
	"github.com/hughbliss/my_toolkit/telemetry/tracer/trace_propagator"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

func MainGateway() (http.Handler, error) {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	defaultGRPCOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			otelgrpc.WithMeterProvider(otel.GetMeterProvider()),
		)),
		grpc.WithStatsHandler(trace_propagator.ClientTracePropagator()),
	}

	/* EXAMPLE_MICROSERVICE */
	if err := someservicev1.RegisterSomeServiceHandlerFromEndpoint(
		ctx, mux, *connectionStringSomeService, defaultGRPCOptions); err != nil {
		return nil, err
	}

	/* AUTH_MICROSERVICE */
	if err := perserv1.RegisterPermissionsServiceHandlerFromEndpoint(
		ctx, mux, *connectionStringAuthService, defaultGRPCOptions); err != nil {
		return nil, err
	}

	return mux, nil
}
