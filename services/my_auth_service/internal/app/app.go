package app

import (
	"context"
	zfg "github.com/chaindead/zerocfg"
	"github.com/hughbliss/my_auth_service/internal/handler"
	"github.com/hughbliss/my_auth_service/internal/usecase"
	"github.com/hughbliss/my_database/pkg/dbauthclient"
	perserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/permissions/v1"
	roleserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/roles/v1"
	"github.com/hughbliss/my_toolkit/cfg"
	"github.com/hughbliss/my_toolkit/grpcerver"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/tracer"
)

var (
	appName = zfg.Str("app_name", "auth_service", "APPNAME")
	appVer  = zfg.Str("app_ver", "local", "APPVER", zfg.Alias("v"))
	env     = zfg.Str("env", "local", "ENV", zfg.Alias("e"))
)

func Run() {
	ctx := context.Background()

	if err := cfg.Init(); err != nil {
		panic(err)
	}
	tracerDown, err := tracer.Init(ctx, *appName, *appVer, *env)
	if err != nil {
		panic(err)
	}
	defer tracerDown()

	reporter.Init(*appName, *appVer, *env, tracer.HookForLogger())

	db, err := dbauthclient.Init(&dbauthclient.Config{Debug: false})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := grpcerver.Init()
	defer s.GracefulStop()

	rolesUsecase := usecase.NewRolesUsecase(db)
	rolesHandler := handler.NewRolesHandler(rolesUsecase)
	roleserv1.RegisterRolesServiceServer(s, rolesHandler)

	permissionsHandler := handler.NewPermissionsHandler()
	perserv1.RegisterPermissionsServiceServer(s, permissionsHandler)

	listener, err := grpcerver.Listener()
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
