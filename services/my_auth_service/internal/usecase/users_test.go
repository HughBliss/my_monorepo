package usecase

import (
	"context"
	"github.com/hughbliss/my_auth_service/internal/dto"
	"github.com/hughbliss/my_database/pkg/dbauthclient"
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func setupUsersTest(t *testing.T) (*UsersUsecase, *dbauth.Client, context.Context) {
	client := dbauthclient.Mock(t)
	usecase := &UsersUsecase{
		rep: reporter.InitReporter("test"),
		db:  client,
	}
	return usecase, client, context.Background()
}

func createTestUserData(t *testing.T, ctx context.Context, client *dbauth.Client) (*dbauth.Domain, *dbauth.Role, *dbauth.User) {
	domain, err := client.Domain.Create().
		SetName("TestDomain").
		Save(ctx)
	require.NoError(t, err)

	role, err := client.Role.Create().
		SetName("TestRole").
		SetDescription("Test Role").
		SetPermissions([]string{"test.permission"}).
		SetDomainID(domain.ID).
		Save(ctx)
	require.NoError(t, err)

	user, err := client.User.Create().
		SetName("TestUser").
		SetEmail("test@example.com").
		SetPasswordHash("hash").
		SetCurrentDomainID(domain.ID).
		Save(ctx)
	require.NoError(t, err)

	return domain, role, user
}

func TestUsersUsecase_AdminGetUsers(t *testing.T) {
	usecase, client, ctx := setupUsersTest(t)
	defer client.Close()

	domain, role, user := createTestUserData(t, ctx, client)

	_, err := client.UserDomain.Create().
		SetUserID(user.ID).
		SetDomainID(domain.ID).
		SetRoleID(role.ID).
		Save(ctx)
	require.NoError(t, err)

	t.Run("успешное получение пользователей", func(t *testing.T) {
		users, err := usecase.AdminGetUsers(ctx)
		require.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 1)
		assert.Equal(t, user.Name, users[0].Name)
	})
}

func TestUsersUsecase_CreateUser(t *testing.T) {
	usecase, client, ctx := setupUsersTest(t)
	defer client.Close()

	domain, _, _ := createTestUserData(t, ctx, client)

	t.Run("успешное создание пользователя", func(t *testing.T) {
		newUser := &dto.User{
			User: dbauth.User{
				Name:            "NewUser",
				Email:           "new@example.com",
				CurrentDomainID: domain.ID,
			},
		}

		result, err := usecase.CreateUser(ctx, newUser)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newUser.Name, result.Name)
		assert.Equal(t, newUser.Email, result.Email)
	})

	t.Run("некорректные данные пользователя", func(t *testing.T) {
		invalidUser := &dto.User{
			User: dbauth.User{
				Name:  "",
				Email: "",
			},
		}

		_, err := usecase.CreateUser(ctx, invalidUser)
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), InvalidUserDataErr.Err().Error())
	})
}

func TestUsersUsecase_UpdateUser(t *testing.T) {
	usecase, client, ctx := setupUsersTest(t)
	defer client.Close()

	domain, _, user := createTestUserData(t, ctx, client)

	t.Run("успешное обновление пользователя", func(t *testing.T) {
		updateUser := &dto.User{
			User: dbauth.User{
				ID:              user.ID,
				Name:            "UpdatedUser",
				Email:           "updated@example.com",
				CurrentDomainID: domain.ID,
			},
		}

		result, err := usecase.UpdateUser(ctx, updateUser)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updateUser.Name, result.Name)
		assert.Equal(t, updateUser.Email, result.Email)
	})

	t.Run("некорректные данные пользователя", func(t *testing.T) {
		invalidUser := &dto.User{
			User: dbauth.User{
				ID:    xid.New(),
				Name:  "",
				Email: "",
			},
		}

		_, err := usecase.UpdateUser(ctx, invalidUser)
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), InvalidUserDataErr.Err().Error())
	})
}

func TestUsersUsecase_DeleteUser(t *testing.T) {
	usecase, client, ctx := setupUsersTest(t)
	defer client.Close()

	_, _, user := createTestUserData(t, ctx, client)

	t.Run("успешное удаление пользователя", func(t *testing.T) {
		err := usecase.DeleteUser(ctx, &dto.User{User: dbauth.User{ID: user.ID}})
		require.NoError(t, err)

		// Проверяем что пользователь действительно удален
		_, err = client.User.Get(ctx, user.ID)
		assert.Error(t, err)
	})

	t.Run("некорректный ID пользователя", func(t *testing.T) {
		err := usecase.DeleteUser(ctx, &dto.User{User: dbauth.User{ID: xid.ID{}}})
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), InvalidUserDataErr.Err().Error())
	})
}

func TestUsersUsecase_AssignUserToDomain(t *testing.T) {
	usecase, client, ctx := setupUsersTest(t)
	defer client.Close()

	domain, role, user := createTestUserData(t, ctx, client)

	t.Run("успешное назначение роли", func(t *testing.T) {
		result, err := usecase.AssignUserToDomain(ctx, user.ID, domain.ID, role.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.Name, result.Name)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		_, err := usecase.AssignUserToDomain(ctx, xid.New(), domain.ID, role.ID)
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), UserNotFoundErr.Err().Error())
	})
}

func TestUsersUsecase_UpdateRole(t *testing.T) {
	usecase, client, ctx := setupUsersTest(t)
	defer client.Close()

	domain, role, user := createTestUserData(t, ctx, client)

	_, err := client.UserDomain.Create().
		SetUserID(user.ID).
		SetDomainID(domain.ID).
		SetRoleID(role.ID).
		Save(ctx)
	require.NoError(t, err)

	newRole, err := client.Role.Create().
		SetName("NewRole").
		SetDomainID(domain.ID).
		SetDescription("NewRole").
		SetPermissions([]string{"read"}).
		Save(ctx)
	require.NoError(t, err)

	t.Run("успешное обновление роли", func(t *testing.T) {
		result, err := usecase.UpdateRole(ctx, user.ID, domain.ID, newRole.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.Name, result.Name)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		_, err := usecase.UpdateRole(ctx, xid.New(), domain.ID, role.ID)
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), UserNotFoundErr.Err().Error())
	})
}
