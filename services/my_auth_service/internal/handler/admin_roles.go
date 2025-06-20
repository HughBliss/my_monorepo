package handler

import (
	"context"
	"errors"
	"github.com/hughbliss/my_auth_service/internal/dto"
	admrolserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/roles/v1"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/rs/xid"
)

func NewAdminRolesHandler(usecase RolesUsecase) admrolserv1.AdminRolesServiceServer {
	return &AdminRolesHandler{
		rep:     reporter.InitReporter("AdminRolesHandler"),
		usecase: usecase,
	}
}

type AdminRolesHandler struct {
	rep     reporter.Reporter
	usecase RolesUsecase
}
type RolesUsecase interface {
	GetDomainsRoles(ctx context.Context) (dto.DomainRolesList, error)
	CreateRole(ctx context.Context, role *dto.Role) (*dto.DomainRoles, error)
	UpdateRole(ctx context.Context, role *dto.Role) (*dto.DomainRoles, error)
	DeleteRole(ctx context.Context, roleID xid.ID) (*dto.DomainRoles, error)
}

func (r AdminRolesHandler) CreateRole(ctx context.Context, request *admrolserv1.CreateRoleRequest) (*admrolserv1.CreateRoleResponse, error) {
	ctx, _, end := r.rep.Start(ctx, "CreateRole")
	defer end()

	domainRoles, err := r.usecase.CreateRole(ctx, dto.RoleFromProto(request.Role))
	if err != nil {
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admrolserv1.CreateRoleResponse{
		DomainRoles: domainRoles.ToProto(),
	}, nil
}

func (r AdminRolesHandler) UpdateRole(ctx context.Context, request *admrolserv1.UpdateRoleRequest) (*admrolserv1.UpdateRoleResponse, error) {
	ctx, _, end := r.rep.Start(ctx, "UpdateRole")
	defer end()

	domainRoles, err := r.usecase.UpdateRole(ctx, dto.RoleFromProto(request.Role))
	if err != nil {
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admrolserv1.UpdateRoleResponse{
		DomainRoles: domainRoles.ToProto(),
	}, nil
}

func (r AdminRolesHandler) DeleteRole(ctx context.Context, request *admrolserv1.DeleteRoleRequest) (*admrolserv1.DeleteRoleResponse, error) {
	ctx, _, end := r.rep.Start(ctx, "DeleteRole")
	defer end()

	roleXID, err := xid.FromString(request.GetRoleId())
	if err != nil {
		return nil, fault.UnhandledError.Err().ToProto()
	}

	domainRoles, err := r.usecase.DeleteRole(ctx, roleXID)
	if err != nil {
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admrolserv1.DeleteRoleResponse{
		DomainRoles: domainRoles.ToProto(),
	}, nil
}

func (r AdminRolesHandler) GetAllRoles(ctx context.Context, _ *admrolserv1.GetAllRolesRequest) (*admrolserv1.GetAllRolesResponse, error) {
	ctx, _, end := r.rep.Start(ctx, "GetAllRoles")
	defer end()

	roles, err := r.usecase.GetDomainsRoles(ctx)
	if err != nil {
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admrolserv1.GetAllRolesResponse{
		DomainRoles: roles.ToProto(),
	}, nil
}
