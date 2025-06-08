package gateway

import (
	"context"
	zfg "github.com/chaindead/zerocfg"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hughbliss/my_protobuf/gen/someservice"
	"github.com/hughbliss/my_toolkit/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
)

var (
	connectionsGroup            = zfg.NewGroup("connection")
	connectionStringSomeService = zfg.Str("some_service", "0.0.0.0:11000", "CONNECTION_SOMESERVICE", zfg.Group(connectionsGroup))

	// example declaring connection strings config
	//connectionStringYetAnotherService = zfg.Str("yet_another_service", "0.0.0.0:11000", "CONNECTION_YETANOTHERSERVICE", zfg.Group(connectionsGroup))
)

func Gateway(ctx context.Context) (http.Handler, error) {
	mux := runtime.NewServeMux()

	defaultGRPCOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(tracer.Provider))),
		grpc.WithStatsHandler(tracer.ClientTraceProvider()),
	}

	if err := someservice.RegisterSomeServiceHandlerFromEndpoint(
		ctx, mux, *connectionStringSomeService, defaultGRPCOptions); err != nil {
		return nil, err
	}

	return mux, nil
}
