package usecase

import (
	"context"
	"github.com/hughbliss/my_auth_service/internal/dto"
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	"github.com/hughbliss/my_database/pkg/gen/dbauth/domain"
	entRole "github.com/hughbliss/my_database/pkg/gen/dbauth/role"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/rs/xid"
)

func NewRolesUsecase(db *dbauth.Client) *RolesUsecase {
	return &RolesUsecase{
		db:  db,
		rep: reporter.InitReporter("RolesUsecase"),
	}
}

// Документация к ошибкам должна быть такой, чтобы она была валидна для вставки в
// ru.yaml и при этом удовлетворяла go-doc. Начало комментария должно
// соответствовать названию сущности, затем двоеточие и тело комментария
// обернутое в двойные кавычки

const (
	RolesGettingDBErr  fault.Code = "RolesGettingDBErr"  // RolesGettingDBErr: "ошибка получения ролей из базы данных"
	RoleCreationDBErr  fault.Code = "RoleCreationDBErr"  // RoleCreationDBErr: "ошибка создания роли в базе данных"
	RoleUpdateDBErr    fault.Code = "RoleUpdateDBErr"    // RoleUpdateDBErr: "ошибка обновления роли в базе данных"
	RoleDeletionDBErr  fault.Code = "RoleDeletionDBErr"  // RoleDeletionDBErr: "ошибка удаления роли из базы данных"
	RoleNotFoundErr    fault.Code = "RoleNotFoundErr"    // RoleNotFoundErr: "роль не найдена"
	DomainNotFoundErr  fault.Code = "DomainNotFoundErr"  // DomainNotFoundErr: "домен не найден"
	InvalidRoleDataErr fault.Code = "InvalidRoleDataErr" // InvalidRoleDataErr: "некорректные данные роли"
)

type RolesUsecase struct {
	rep reporter.Reporter
	db  *dbauth.Client
}

func (r RolesUsecase) GetDomainsRoles(ctx context.Context) (dto.DomainRolesList, error) {
	ctx, log, end := r.rep.Start(ctx, "GetDomainsRoles")
	defer end()

	all, err := r.db.Domain.Query().WithRoles().All(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to query all roles")
		return nil, RolesGettingDBErr.Err()
	}

	return dto.DomainRolesListFromEnt(all), nil
}

func (r RolesUsecase) GetDomainRoles(ctx context.Context, domainId xid.ID) (*dto.DomainRoles, error) {
	ctx, log, end := r.rep.Start(ctx, "GetDomainRoles")
	defer end()

	domain, err := r.db.Domain.Query().Where(domain.ID(domainId)).WithRoles().Only(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to query single domain with roles")
		return nil, RolesGettingDBErr.Err()
	}

	return dto.DomainRolesFromEnt(domain), nil

}

func (r RolesUsecase) CreateRole(ctx context.Context, role *dto.Role) (*dto.DomainRoles, error) {
	ctx, log, end := r.rep.Start(ctx, "CreateRole")
	defer end()

	if role.Name == "" || role.DomainId.IsNil() {
		log.Warn().Msg("invalid role data")
		return nil, InvalidRoleDataErr.Err()
	}

	domain, err := r.db.Domain.Get(ctx, role.DomainId)
	if err != nil {
		log.Err(err).Stack().Msg("failed to get domain")
		return nil, DomainNotFoundErr.Err()
	}

	if _, err := r.db.Role.Create().
		SetName(role.Name).
		SetDescription(role.Description).
		SetPermissions(role.Permissions).
		SetDomainID(domain.ID).
		Save(ctx); err != nil {
		log.Err(err).Stack().Msg("failed to create role")
		return nil, RoleCreationDBErr.Err()
	}

	return r.GetDomainRoles(ctx, domain.ID)
}

func (r RolesUsecase) UpdateRole(ctx context.Context, role *dto.Role) (*dto.DomainRoles, error) {
	ctx, log, end := r.rep.Start(ctx, "UpdateRole")
	defer end()

	if role.Name == "" || role.DomainId.IsNil() || role.ID.IsNil() {
		log.Warn().Msg("invalid role data")
		return nil, InvalidRoleDataErr.Err()
	}

	existingRole, err := r.db.Role.Query().
		Where(entRole.DomainID(role.DomainId)).
		Where(entRole.ID(role.ID)).First(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to find role")
		return nil, RoleNotFoundErr.Err()
	}
	if _, err := r.db.Role.UpdateOne(existingRole).
		SetName(role.Name).
		SetDescription(role.Description).
		SetPermissions(role.Permissions).
		Save(ctx); err != nil {
		log.Err(err).Stack().Msg("failed to update role")
		return nil, RoleUpdateDBErr.Err()
	}

	return r.GetDomainRoles(ctx, role.DomainId)
}

func (r RolesUsecase) DeleteRole(ctx context.Context, roleID xid.ID) (*dto.DomainRoles, error) {
	ctx, log, end := r.rep.Start(ctx, "DeleteRole")
	defer end()

	role, err := r.db.Role.Get(ctx, roleID)
	if err != nil {
		log.Err(err).Stack().Msg("failed to find role")
		return nil, RoleNotFoundErr.Err()
	}

	domainID := role.DomainID

	if err = r.db.Role.DeleteOne(role).Exec(ctx); err != nil {
		log.Err(err).Stack().Msg("failed to delete role")
		return nil, RoleDeletionDBErr.Err()
	}
	return r.GetDomainRoles(ctx, domainID)
}
