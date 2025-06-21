package service

import (
	"github.com/hughbliss/my_gateway/internal/gateway"
	authnv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/authn/v1"
	"google.golang.org/grpc"
)

func NewAuthenticationService() (authnv1.AuthenticationServiceClient, error) {
	connection, err := grpc.NewClient(*gateway.ConnectionStringAuthService, gateway.DefaultGRPCOptions...)
	if err != nil {
		return nil, err
	}
	return authnv1.NewAuthenticationServiceClient(connection), nil
}
