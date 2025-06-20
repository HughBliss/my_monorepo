package app

import (
	"context"
	"fmt"
	zfg "github.com/chaindead/zerocfg"
	"github.com/hughbliss/my_gateway/internal/gateway"
	"github.com/hughbliss/my_toolkit/cfg"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/telemetry"
	"github.com/hughbliss/my_toolkit/telemetry/tracer"
	traceExporter "github.com/hughbliss/my_toolkit/telemetry/tracer/exporter/jaeger"
	"github.com/hughbliss/my_toolkit/telemetry/tracer/trace_middleware"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
)

var (
	listenGroup = zfg.NewGroup("listen")
	listenHost  = zfg.Str("host", "0.0.0.0", "LISTEN_HOST", zfg.Group(listenGroup))
	listenPort  = zfg.Uint32("port", 8080, "LISTEN_PORT", zfg.Group(listenGroup))
)

var (
	appName = zfg.Str("app_name", "my_gateway", "APPNAME")
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

	if err := cfg.Init(); err != nil {
		panic(err)
	}

	telemetryDown := initTelemetry()
	defer telemetryDown()

	reporter.Init(*appName, *appVer, *env, trace_middleware.HookForLogger())

	e := echo.New()
	registerMiddleware(e)

	mainGroup := e.Group("")
	mainGatewayHandler, err := gateway.MainGateway()
	if err != nil {
		panic(err)
	}
	mainGroup.Any("/*", echo.WrapHandler(mainGatewayHandler))

	adminGroup := e.Group("/admin")
	adminGatewayHandler, err := gateway.AdminGateway()
	if err != nil {
		panic(err)
	}
	adminGroup.Any("/*", echo.WrapHandler(adminGatewayHandler))

	if err := e.Start(fmt.Sprintf("%s:%d", *listenHost, *listenPort)); err != nil {
		panic(err)
	}
}

func registerMiddleware(e *echo.Echo) {
	e.Use(echoMiddleware.LoggerWithConfig(echoMiddleware.LoggerConfig{
		Format: "${status} ${method} ${uri}",
		Output: log.With().Str("level", "info").Str("component", "echo").Logger(),
	}))
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.Gzip())
	e.Use(echoMiddleware.BodyLimit("2M"))
	e.Use(otelecho.Middleware(*appName,
		otelecho.WithTracerProvider(otel.GetTracerProvider()),
	))
	e.Use(trace_middleware.AddTraceIDToResponse)

}
