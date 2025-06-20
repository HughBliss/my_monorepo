package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	admrolserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/roles/v1"
	admusrserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/users/v1"
	"github.com/hughbliss/my_toolkit/telemetry/tracer/trace_propagator"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

func AdminGateway() (http.Handler, error) {
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

	/* AUTH_MICROSERVICE */
	if err := admrolserv1.RegisterAdminRolesServiceHandlerFromEndpoint(
		ctx, mux, *connectionStringAuthService, defaultGRPCOptions); err != nil {
		return nil, err
	}
	if err := admusrserv1.RegisterAdminUsersServiceHandlerFromEndpoint(
		ctx, mux, *connectionStringAuthService, defaultGRPCOptions); err != nil {
		return nil, err
	}

	return mux, nil
}
