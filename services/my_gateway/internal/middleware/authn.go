package middleware

import (
	"context"
	"github.com/hughbliss/my_protobuf/go/pkg/gen/acman"
	authnv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/authn/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

func AuthInterceptor(service authnv1.AuthenticationServiceClient) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		requiredPermission, ok := acman.MethodPermissionMap[method]
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authHeaders := md.Get("Authorization")
		if len(authHeaders) != 1 {
			return status.Error(codes.Unauthenticated, "authorization header is invalid")
		}

		accessToken := strings.TrimPrefix(authHeaders[0], "Bearer ")

		userMeta, err := service.Authorize(ctx, &authnv1.AuthorizeRequest{
			AccessToken: accessToken,
		})
		if err != nil {
			return err
		}

		for _, permission := range userMeta.Permissions {
			if permission == requiredPermission.Alias {
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}

		return status.Error(codes.PermissionDenied, "у вас нет доступа: "+requiredPermission.Description)
	}
}
