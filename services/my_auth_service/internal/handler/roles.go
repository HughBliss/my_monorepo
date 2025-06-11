package handler

import (
	"context"
	"errors"
	"github.com/hughbliss/my_auth_service/internal/dto"
	roleserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/roles/v1"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
)

func NewRolesHandler(usecase RolesUsecase) roleserv1.RolesServiceServer {
	return &RolesHandler{
		rep:     reporter.InitReporter("RolesHandler"),
		usecase: usecase,
	}
}

type RolesHandler struct {
	rep     reporter.Reporter
	usecase RolesUsecase
}

type RolesUsecase interface {
	GetRoles(ctx context.Context) (dto.DomainRolesList, error)
}

func (r RolesHandler) GetAllRoles(ctx context.Context, _ *roleserv1.GetAllRolesRequest) (*roleserv1.GetAllRolesResponse, error) {
	ctx, _, end := r.rep.Start(ctx, "GetAllRoles")
	defer end()

	roles, err := r.usecase.GetRoles(ctx)
	if err != nil {
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &roleserv1.GetAllRolesResponse{
		DomainRoles: roles.ToProto(),
	}, nil
}
