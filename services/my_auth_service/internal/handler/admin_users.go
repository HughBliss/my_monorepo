package handler

import (
	"context"
	"errors"
	"github.com/hughbliss/my_auth_service/internal/dto"
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	admusrserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/users/v1"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/rs/xid"
)

func NewAdminUserHandler(uc UsersUsecase) admusrserv1.AdminUsersServiceServer {
	return &AdminUserHandler{
		rep: reporter.InitReporter("AdminUserHandler"),
		uc:  uc,
	}
}

type AdminUserHandler struct {
	rep reporter.Reporter
	uc  UsersUsecase
}

type UsersUsecase interface {
	AdminGetUsers(ctx context.Context) (dto.UserList, error)
	CreateUser(ctx context.Context, user *dto.User) (*dto.User, error)
	UpdateUser(ctx context.Context, user *dto.User) (*dto.User, error)
	DeleteUser(ctx context.Context, user *dto.User) error

	AssignUserToDomain(ctx context.Context, userID, domainID, roleID xid.ID) (*dto.User, error)
	RemoveUserFromDomain(ctx context.Context, userID, domainID xid.ID) (*dto.User, error)
	UpdateRole(ctx context.Context, userID, domainID, roleID xid.ID) (*dto.User, error)
}

func (a AdminUserHandler) CreateUser(ctx context.Context, request *admusrserv1.CreateUserRequest) (*admusrserv1.CreateUserResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "CreateUser")
	defer end()

	u := new(dto.User).FromProto(request.GetUser())

	user, err := a.uc.CreateUser(ctx, u)
	if err != nil {
		log.Error().Err(err).Msg("CreateUser")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.CreateUserResponse{
		UserDomains: user.ToProto(),
	}, nil

}
func (a AdminUserHandler) AssignUserToDomain(ctx context.Context, request *admusrserv1.AssignUserToDomainRequest) (*admusrserv1.AssignUserToDomainResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "AssignUser")
	defer end()

	userID, err := xid.FromString(request.UserId)
	if err != nil {
		log.Error().Err(err).Msg("AssignUser")
		return nil, fault.UnhandledError.Err().ToProto()
	}
	domainID, err := xid.FromString(request.DomainId)
	if err != nil {
		log.Error().Err(err).Msg("AssignUser")
		return nil, fault.UnhandledError.Err().ToProto()
	}
	roleID, err := xid.FromString(request.RoleId)
	if err != nil {
		log.Error().Err(err).Msg("AssignUser")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	user, err := a.uc.AssignUserToDomain(ctx, userID, domainID, roleID)
	if err != nil {
		log.Error().Err(err).Msg("AssignUser")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.AssignUserToDomainResponse{UserDomains: user.ToProto()}, nil

}
func (a AdminUserHandler) UpdateUser(ctx context.Context, request *admusrserv1.UpdateUserRequest) (*admusrserv1.UpdateUserResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "UpdateUser")
	defer end()

	u := new(dto.User).FromProto(request.GetUser())

	user, err := a.uc.UpdateUser(ctx, u)
	if err != nil {
		log.Error().Err(err).Msg("UpdateUser")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.UpdateUserResponse{
		UserDomains: user.ToProto(),
	}, nil
}

func (a AdminUserHandler) DeleteUser(ctx context.Context, request *admusrserv1.DeleteUserRequest) (*admusrserv1.DeleteUserResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "DeleteUser")
	defer end()

	userID, err := xid.FromString(request.UserId)
	if err != nil {
		log.Error().Err(err).Msg("DeleteUser")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	err = a.uc.DeleteUser(ctx, &dto.User{User: dbauth.User{ID: userID}})
	if err != nil {
		log.Error().Err(err).Msg("DeleteUser")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.DeleteUserResponse{}, nil
}

func (a AdminUserHandler) GetUsers(ctx context.Context, _ *admusrserv1.GetUsersRequest) (*admusrserv1.GetUsersResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "GetUsers")
	defer end()

	users, err := a.uc.AdminGetUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("GetUsers")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.GetUsersResponse{
		Users: users.ToProto(),
	}, nil
}

func (a AdminUserHandler) RemoveUserFromDomain(ctx context.Context, request *admusrserv1.RemoveUserFromDomainRequest) (*admusrserv1.RemoveUserFromDomainResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "RemoveUserFromDomain")
	defer end()

	userID, err := xid.FromString(request.UserId)
	if err != nil {
		log.Error().Err(err).Msg("RemoveUserFromDomain")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	domainID, err := xid.FromString(request.DomainId)
	if err != nil {
		log.Error().Err(err).Msg("RemoveUserFromDomain")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	user, err := a.uc.RemoveUserFromDomain(ctx, userID, domainID)
	if err != nil {
		log.Error().Err(err).Msg("RemoveUserFromDomain")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.RemoveUserFromDomainResponse{
		UserDomains: user.ToProto(),
	}, nil
}

func (a AdminUserHandler) UpdateUserDomainRole(ctx context.Context, request *admusrserv1.UpdateUserDomainRoleRequest) (*admusrserv1.UpdateUserDomainRoleResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "UpdateUserDomainRole")
	defer end()

	userID, err := xid.FromString(request.UserId)
	if err != nil {
		log.Error().Err(err).Msg("UpdateUserDomainRole")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	domainID, err := xid.FromString(request.DomainId)
	if err != nil {
		log.Error().Err(err).Msg("UpdateUserDomainRole")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	roleID, err := xid.FromString(request.RoleId)
	if err != nil {
		log.Error().Err(err).Msg("UpdateUserDomainRole")
		return nil, fault.UnhandledError.Err().ToProto()
	}

	user, err := a.uc.UpdateRole(ctx, userID, domainID, roleID)
	if err != nil {
		log.Error().Err(err).Msg("UpdateUserDomainRole")
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &admusrserv1.UpdateUserDomainRoleResponse{
		UserDomains: user.ToProto(),
	}, nil
}
