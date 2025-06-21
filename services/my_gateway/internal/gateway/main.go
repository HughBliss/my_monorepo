package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	authnv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/authn/v1"
	perserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/permissions/v1"
	someservicev1 "github.com/hughbliss/my_protobuf/go/pkg/gen/someservice/v1"
	"google.golang.org/grpc"
	"net/http"
)

func MainGateway(authInterceptor grpc.UnaryClientInterceptor) (http.Handler, error) {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	withAuth := []grpc.DialOption{
		grpc.WithUnaryInterceptor(authInterceptor),
	}
	withAuth = append(withAuth, DefaultGRPCOptions...)

	/* EXAMPLE_MICROSERVICE */
	if err := someservicev1.RegisterSomeServiceHandlerFromEndpoint(
		ctx, mux, *ConnectionStringSomeService, withAuth); err != nil {
		return nil, err
	}

	/* AUTH_MICROSERVICE */
	if err := perserv1.RegisterPermissionsServiceHandlerFromEndpoint(
		ctx, mux, *ConnectionStringAuthService, withAuth); err != nil {
		return nil, err
	}
	if err := authnv1.RegisterAuthenticationServiceHandlerFromEndpoint(
		ctx, mux, *ConnectionStringAuthService, DefaultGRPCOptions); err != nil {
		return nil, err
	}

	return mux, nil
}
