package usecase

import (
	"context"
	"github.com/hughbliss/my_auth_service/internal/dto"
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
)

func NewRolesUsecase(db *dbauth.Client) *RolesUsecase {
	return &RolesUsecase{
		db:  db,
		rep: reporter.InitReporter("RolesUsecase"),
	}
}

const (
	RolesGettingDBErr fault.Code = "RolesGettingDBErr" // RolesGettingDBErr: "ошибка получение ролей из базы данных"
)

type RolesUsecase struct {
	rep reporter.Reporter
	db  *dbauth.Client
}

func (r RolesUsecase) GetRoles(ctx context.Context) (dto.DomainRolesList, error) {
	ctx, log, end := r.rep.Start(ctx, "GetRoles")
	defer end()

	all, err := r.db.Domain.Query().WithRoles().All(ctx)
	if err != nil {
		log.Err(err).Stack().Msg("failed to query all roles")
		return nil, RolesGettingDBErr.Err()
	}

	return dto.DomainRolesListFromEnt(all), nil

}
