package app

import (
	"context"
	zfg "github.com/chaindead/zerocfg"
	"github.com/hughbliss/my_auth_service/internal/handler"
	"github.com/hughbliss/my_auth_service/internal/usecase"
	"github.com/hughbliss/my_database/pkg/dbauthclient"
	admrolserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/roles/v1"
	admusrserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/admin/users/v1"
	perserv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/permissions/v1"
	"github.com/hughbliss/my_toolkit/cfg"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/grpcerver"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/telemetry"
	"github.com/hughbliss/my_toolkit/telemetry/tracer"
	traceExporter "github.com/hughbliss/my_toolkit/telemetry/tracer/exporter/jaeger"
	"github.com/hughbliss/my_toolkit/telemetry/tracer/trace_middleware"
)

var (
	appName = zfg.Str("app_name", "auth_service", "APPNAME")
	appVer  = zfg.Str("app_ver", "local", "APPVER", zfg.Alias("v"))
	env     = zfg.Str("env", "local", "ENV", zfg.Alias("e"))
)

func initTelemetry() func() {
	ctx := context.Background()
	resourceMeta := telemetry.ResourceMeta(*appName, *appVer, *env)

	jaegerExporter, err := traceExporter.Jaeger(ctx)
	if err != nil {
		panic(err)
	}

	if err := fault.InitLocales("./locales/ru.yaml"); err != nil {
		panic(err)
	}

	//otlpMeter, err := meterExporter.OTLPMeter(ctx)
	//if err != nil {
	//	panic(err)
	//}

	tracerDown := tracer.Init(ctx, resourceMeta, jaegerExporter)
	//meterDown := meter.Init(ctx, resourceMeta, otlpMeter)

	return func() {
		//meterDown()
		tracerDown()
	}
}

func Run() {

	if err := cfg.Init(); err != nil {
		panic(err)
	}

	lelemetryDown := initTelemetry()
	defer lelemetryDown()

	reporter.Init(*appName, *appVer, *env, trace_middleware.HookForLogger())

	db, err := dbauthclient.Init(&dbauthclient.Config{Debug: false})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := grpcerver.Init()
	defer s.GracefulStop()

	// REPOSITORIES

	// USECASES
	rolesUsecase := usecase.NewRolesUsecase(db)
	usersUsecase := usecase.NewUsersUsecase(db)

	// HANDLERS
	adminRolesHandler := handler.NewAdminRolesHandler(rolesUsecase)
	admrolserv1.RegisterAdminRolesServiceServer(s, adminRolesHandler)

	adminUserHandler := handler.NewAdminUserHandler(usersUsecase)
	admusrserv1.RegisterAdminUsersServiceServer(s, adminUserHandler)

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
