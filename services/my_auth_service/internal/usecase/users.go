package usecase

import (
	"context"
	"github.com/hughbliss/my_auth_service/internal/dto"
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	entUser "github.com/hughbliss/my_database/pkg/gen/dbauth/user"
	"github.com/hughbliss/my_database/pkg/gen/dbauth/userdomain"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/rs/xid"
)

// Документация к ошибкам должна быть такой, чтобы она была валидна для вставки в
// ru.yaml и при этом удовлетворяла go-doc. Начало комментария должно
// соответствовать названию сущности, затем двоеточие и тело комментария
// обернутое в двойные кавычки

const (
	UsersGettingDBErr  fault.Code = "UsersGettingDBErr"  // UsersGettingDBErr: "ошибка получения пользователей из базы данных"
	UserCreationDBErr  fault.Code = "UserCreationDBErr"  // UserCreationDBErr: "ошибка создания пользователя в базе данных"
	UserUpdateDBErr    fault.Code = "UserUpdateDBErr"    // UserUpdateDBErr: "ошибка обновления пользователя в базе данных"
	UserDeletionDBErr  fault.Code = "UserDeletionDBErr"  // UserDeletionDBErr: "ошибка удаления пользователя из базы данных"
	UserNotFoundErr    fault.Code = "UserNotFoundErr"    // UserNotFoundErr: "пользователь не найден"
	InvalidUserDataErr fault.Code = "InvalidUserDataErr" // InvalidUserDataErr: "некорректные данные пользователя"
)

func NewUsersUsecase(db *dbauth.Client) *UsersUsecase {
	return &UsersUsecase{
		rep: reporter.InitReporter("UsersUsecase"),
		db:  db,
	}
}

type UsersUsecase struct {
	rep reporter.Reporter
	db  *dbauth.Client
}

func (u UsersUsecase) AdminGetUsers(ctx context.Context) (dto.UserList, error) {
	ctx, log, end := u.rep.Start(ctx, "AdminGetUsers")
	defer end()

	users, err := u.db.User.Query().WithUserDomain(func(userDomainQuery *dbauth.UserDomainQuery) {
		userDomainQuery.WithDomain().WithRole()
	}).All(ctx)
	if err != nil {
		log.Error().Err(err).Stack().Msg("AdminGetUsers")
		return nil, UsersGettingDBErr.Err()
	}

	return new(dto.UserList).FromEnt(users), nil

}
func (u UsersUsecase) CreateUser(ctx context.Context, user *dto.User) (*dto.User, error) {
	ctx, log, end := u.rep.Start(ctx, "CreateUser")
	defer end()

	if user.Name == "" || user.Email == "" {
		log.Warn().Msg("invalid user data")
		return nil, InvalidUserDataErr.Err()
	}

	created, err := u.db.User.Create().
		SetName(user.Name).
		SetEmail(user.Email).
		SetCurrentDomainID(user.CurrentDomainID).
		SetPasswordHash("todo: password hash").
		Save(ctx)

	if err != nil {
		log.Err(err).Stack().Msg("failed to create user")
		return nil, UserCreationDBErr.Err()
	}

	return new(dto.User).FromEnt(created), nil
}

func (u UsersUsecase) UpdateUser(ctx context.Context, user *dto.User) (*dto.User, error) {
	ctx, log, end := u.rep.Start(ctx, "UpdateUser")
	defer end()

	if user.ID.IsNil() || user.Name == "" || user.Email == "" {
		log.Warn().Msg("invalid user data")
		return nil, InvalidUserDataErr.Err()
	}

	updated, err := u.db.User.UpdateOneID(user.ID).
		SetName(user.Name).
		SetEmail(user.Email).
		SetCurrentDomainID(user.CurrentDomainID).
		Save(ctx)

	if err != nil {
		log.Err(err).Stack().Msg("failed to update user")
		return nil, UserUpdateDBErr.Err()
	}

	return new(dto.User).FromEnt(updated), nil
}

func (u UsersUsecase) DeleteUser(ctx context.Context, user *dto.User) error {
	ctx, log, end := u.rep.Start(ctx, "DeleteUser")
	defer end()

	if user.ID.IsNil() {
		log.Warn().Msg("invalid user id")
		return InvalidUserDataErr.Err()
	}

	err := u.db.User.DeleteOneID(user.ID).Exec(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to delete user")
		return UserDeletionDBErr.Err()
	}

	return nil
}

func (u UsersUsecase) AssignUserToDomain(ctx context.Context, userID, domainID, roleID xid.ID) (*dto.User, error) {
	ctx, log, end := u.rep.Start(ctx, "AssignUserToDomain")
	defer end()

	user, err := u.db.User.Query().Where(entUser.ID(userID)).Only(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to find user")
		return nil, UserNotFoundErr.Err()
	}

	_, err = u.db.UserDomain.Create().
		SetUserID(userID).
		SetDomainID(domainID).
		SetRoleID(roleID).
		Save(ctx)

	if err != nil {
		log.Err(err).Stack().Msg("failed to assign user to domain")
		return nil, UserUpdateDBErr.Err()
	}

	return new(dto.User).FromEnt(user), nil
}

func (u UsersUsecase) RemoveUserFromDomain(ctx context.Context, userID, domainID xid.ID) (*dto.User, error) {
	ctx, log, end := u.rep.Start(ctx, "RemoveUserFromDomain")
	defer end()

	user, err := u.db.User.Query().Where(entUser.ID(userID)).Only(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to find user")
		return nil, UserNotFoundErr.Err()
	}

	_, err = u.db.UserDomain.Delete().
		Where(userdomain.UserID(userID)).
		Where(userdomain.DomainID(domainID)).
		Exec(ctx)

	if err != nil {
		log.Err(err).Stack().Msg("failed to remove user from domain")
		return nil, UserUpdateDBErr.Err()
	}

	return new(dto.User).FromEnt(user), nil
}

func (u UsersUsecase) UpdateRole(ctx context.Context, userID, domainID, roleID xid.ID) (*dto.User, error) {
	ctx, log, end := u.rep.Start(ctx, "UpdateRole")
	defer end()

	user, err := u.db.User.Query().Where(entUser.ID(userID)).Only(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to find user")
		return nil, UserNotFoundErr.Err()
	}

	_, err = u.db.UserDomain.Update().
		Where(userdomain.UserID(userID)).
		Where(userdomain.DomainID(domainID)).
		SetRoleID(roleID).
		Save(ctx)

	if err != nil {
		log.Err(err).Stack().Msg("failed to update user role")
		return nil, UserUpdateDBErr.Err()
	}

	return new(dto.User).FromEnt(user), nil
}
