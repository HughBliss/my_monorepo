package handler

import (
	"context"
	"github.com/hughbliss/my_protobuf/go/pkg/gen/acman"
	perserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/permissions/v1"

	"github.com/hughbliss/my_toolkit/reporter"
)

func NewPermissionsHandler() perserv1.PermissionsServiceServer {
	return &PermissionsHandler{
		rep: reporter.InitReporter("PermissionsHandler"),
	}
}

type PermissionsHandler struct {
	rep reporter.Reporter
}

func (p PermissionsHandler) GetAllPermissions(ctx context.Context, request *perserv1.GetAllPermissionsRequest) (*perserv1.GetAllPermissionsResponse, error) {
	ctx, _, end := p.rep.Start(ctx, "GetAllPermissions")
	defer end()

	return &perserv1.GetAllPermissionsResponse{
		Permissions: acman.Permissions,
	}, nil
}
