package app

import (
	"context"
	"fmt"
	zfg "github.com/chaindead/zerocfg"
	zfgEnv "github.com/chaindead/zerocfg/env"
	zfgYaml "github.com/chaindead/zerocfg/yaml"
	someservicev1 "github.com/hughbliss/my_protobuf/go/pkg/gen/someservice/v1"
	"github.com/hughbliss/my_service/internal/handler"
	"github.com/hughbliss/my_service/internal/usecase"
	"github.com/hughbliss/my_toolkit/grpcerver"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/tracer"
)

var (
	configYamlPath = zfg.Str("cfg_yaml_path", "./config.yaml", "CFGYAMLPATH", zfg.Alias("c"))

	appName = zfg.Str("app_name", "my_service", "APPNAME")
	appVer  = zfg.Str("app_ver", "0.0.1", "APPVER", zfg.Alias("v"))
	env     = zfg.Str("env", "local", "ENV", zfg.Alias("e"))
)

func Run() {
	ctx := context.Background()

	if err := zfg.Parse(zfgEnv.New(), zfgYaml.New(configYamlPath)); err != nil {
		panic(err)
	}
	fmt.Println("starting with config\n", zfg.Show())

	shutdown, err := tracer.Init(ctx, *appName, *appVer, *env)
	if err != nil {
		panic(err)
	}
	defer shutdown()

	reporter.Init(*appName, *appVer, *env, tracer.HookForLogger())

	s := grpcerver.Init()
	defer s.GracefulStop()

	someUsecase := usecase.NewSomeUsecase()

	someServiceHandler := handler.NewSomeServiceHandler(someUsecase)

	someservicev1.RegisterSomeServiceServer(s, someServiceHandler)

	listener, err := grpcerver.Listener()
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	if err := s.Serve(listener); err != nil {
		panic(err)
	}

}
