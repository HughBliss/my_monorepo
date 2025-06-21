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
	"github.com/hughbliss/my_toolkit/telemetry"
	"github.com/hughbliss/my_toolkit/telemetry/tracer"
	traceExporter "github.com/hughbliss/my_toolkit/telemetry/tracer/exporter/jaeger"
	"github.com/hughbliss/my_toolkit/telemetry/tracer/trace_middleware"
)

var (
	configYamlPath = zfg.Str("cfg_yaml_path", "./config.yaml", "CFGYAMLPATH", zfg.Alias("c"))

	appName = zfg.Str("app_name", "my_service", "APPNAME")
	appVer  = zfg.Str("app_ver", "0.0.1", "APPVER", zfg.Alias("v"))
	env     = zfg.Str("env", "local", "ENV", zfg.Alias("e"))
)

func initTelemetry() func() {
	ctx := context.Background()
	resourceMeta := telemetry.ResourceMeta(*appName, *appVer, *env)

	jaegerExporter, err := traceExporter.Jaeger(ctx)
	if err != nil {
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

	if err := zfg.Parse(zfgEnv.New(), zfgYaml.New(configYamlPath)); err != nil {
		panic(err)
	}
	fmt.Println("starting with config\n", zfg.Show())

	lelemetryDown := initTelemetry()
	defer lelemetryDown()

	reporter.Init(*appName, *appVer, *env, trace_middleware.HookForLogger())

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
