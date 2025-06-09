package app

import (
	"context"
	"fmt"
	zfg "github.com/chaindead/zerocfg"
	zfgEnv "github.com/chaindead/zerocfg/env"
	zfgYaml "github.com/chaindead/zerocfg/yaml"
	"github.com/hughbliss/my_protobuf/gen/someservice"
	"github.com/hughbliss/my_service/internal/handler"
	"github.com/hughbliss/my_service/internal/server"
	"github.com/hughbliss/my_service/internal/usecase"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/tracer"
	"net"
)

var (
	listenGroup = zfg.NewGroup("listen")
	listenHost  = zfg.Str("host", "0.0.0.0", "LISTEN_HOST", zfg.Group(listenGroup))
	listenPort  = zfg.Uint32("port", 11000, "LISTEN_PORT", zfg.Group(listenGroup))
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

	reporter.Init(tracer.HookForLogger())

	s := server.Init()
	defer s.GracefulStop()

	someUsecase := usecase.NewSomeUsecase()

	someServiceHandler := handler.NewSomeServiceHandler(someUsecase)

	someservice.RegisterSomeServiceServer(s, someServiceHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *listenHost, *listenPort))
	if err != nil {
		panic(err)
	}

	if err := s.Serve(listener); err != nil {
		panic(err)
	}

}
