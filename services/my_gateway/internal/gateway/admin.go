package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	admrolserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/roles/v1"
	admusrserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/users/v1"
	"google.golang.org/grpc"
	"net/http"
)

func AdminGateway(options ...grpc.DialOption) (http.Handler, error) {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	var opts []grpc.DialOption
	opts = append(opts, DefaultGRPCOptions...)
	for _, option := range options {
		opts = append(opts, option)
	}

	/* AUTH_MICROSERVICE */
	if err := admrolserv1.RegisterAdminRolesServiceHandlerFromEndpoint(
		ctx, mux, *ConnectionStringAuthService, opts); err != nil {
		return nil, err
	}
	if err := admusrserv1.RegisterAdminUsersServiceHandlerFromEndpoint(
		ctx, mux, *ConnectionStringAuthService, opts); err != nil {
		return nil, err
	}

	return mux, nil
}
