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

func setupRolesTest(t *testing.T) (*RolesUsecase, *dbauth.Client, context.Context) {
	client := dbauthclient.Mock(t)
	usecase := &RolesUsecase{
		rep: reporter.InitReporter("test"),
		db:  client,
	}
	return usecase, client, context.Background()
}

func createTestDomain(t *testing.T, ctx context.Context, client *dbauth.Client) *dbauth.Domain {
	domain, err := client.Domain.Create().
		SetName("TestDomain").
		Save(ctx)
	require.NoError(t, err)
	return domain
}

func TestRolesUsecase_CreateRole(t *testing.T) {
	usecase, client, ctx := setupRolesTest(t)
	defer client.Close()
	domain := createTestDomain(t, ctx, client)

	t.Run("успешное создание роли", func(t *testing.T) {
		role := &dto.Role{
			Name:        "TestRole",
			Description: "Test Description",
			Permissions: []string{"perm1", "perm2"},
			DomainId:    domain.ID,
		}

		result, err := usecase.CreateRole(ctx, role)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.Name, result.DomainName)
		assert.Len(t, result.Roles, 1)
		assert.Equal(t, role.Name, result.Roles[0].Name)
	})

	t.Run("домен не найден", func(t *testing.T) {
		role := &dto.Role{
			Name:        "TestRole",
			Description: "Test Description",
			Permissions: []string{"perm1"},
			DomainId:    xid.New(),
		}

		_, err := usecase.CreateRole(ctx, role)
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), DomainNotFoundErr.Err().Error())
	})
}

func TestRolesUsecase_UpdateRole(t *testing.T) {
	usecase, client, ctx := setupRolesTest(t)
	defer client.Close()
	domain := createTestDomain(t, ctx, client)

	role, err := client.Role.Create().
		SetName("InitialRole").
		SetDescription("Initial Description").
		SetPermissions([]string{"perm1"}).
		SetDomainID(domain.ID).
		Save(ctx)
	require.NoError(t, err)

	t.Run("успешное обновление роли", func(t *testing.T) {
		updateRole := &dto.Role{
			ID:          role.ID,
			Name:        "UpdatedRole",
			Description: "Updated Description",
			Permissions: []string{"perm1", "perm2"},
			DomainId:    domain.ID,
		}

		result, err := usecase.UpdateRole(ctx, updateRole)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updateRole.Name, result.Roles[0].Name)
		assert.Equal(t, updateRole.Description, result.Roles[0].Description)
	})

	t.Run("роль не найдена", func(t *testing.T) {
		updateRole := &dto.Role{
			ID:          xid.New(),
			Name:        "UpdatedRole",
			Description: "Updated Description",
			DomainId:    domain.ID,
		}

		_, err := usecase.UpdateRole(ctx, updateRole)
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), RoleNotFoundErr.Err().Error())
	})
}

func TestRolesUsecase_DeleteRole(t *testing.T) {
	usecase, client, ctx := setupRolesTest(t)
	defer client.Close()
	domain := createTestDomain(t, ctx, client)

	role, err := client.Role.Create().
		SetName("RoleToDelete").
		SetDescription("To be deleted").
		SetPermissions([]string{"perm1", "perm2"}).
		SetDomainID(domain.ID).
		Save(ctx)
	require.NoError(t, err)

	t.Run("успешное удаление роли", func(t *testing.T) {
		result, err := usecase.DeleteRole(ctx, role.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Roles)

		// Проверяем что роль действительно удалена
		_, err = client.Role.Get(ctx, role.ID)
		assert.Error(t, err)
	})

	t.Run("роль не найдена", func(t *testing.T) {
		_, err := usecase.DeleteRole(ctx, xid.New())
		assert.Error(t, err)
		f := new(fault.Fault)
		assert.ErrorAs(t, err, &f)
		assert.Equal(t, f.Error(), RoleNotFoundErr.Err().Error())
	})
}

func TestRolesUsecase_GetDomainsRoles(t *testing.T) {
	usecase, client, ctx := setupRolesTest(t)
	defer client.Close()
	domain := createTestDomain(t, ctx, client)

	_, err := client.Role.Create().
		SetName("TestRole1").
		SetDescription("Test Description 1").
		SetPermissions([]string{"perm1"}).
		SetDomainID(domain.ID).
		Save(ctx)
	require.NoError(t, err)

	t.Run("успешное получение ролей", func(t *testing.T) {
		result, err := usecase.GetDomainsRoles(ctx)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, domain.Name, result[0].DomainName)
		assert.Len(t, result[0].Roles, 1)
	})
}
