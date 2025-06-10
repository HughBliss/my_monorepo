package app

import (
	"context"
	"fmt"
	zfg "github.com/chaindead/zerocfg"
	"github.com/hughbliss/my_gateway/internal/gateway"
	"github.com/hughbliss/my_toolkit/cfg"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/tracer"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
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

func Run() {

	ctx := context.Background()

	if err := cfg.Init(); err != nil {
		panic(err)
	}

	shutdown, err := tracer.Init(ctx, *appName, *appVer, *env)
	if err != nil {
		panic(err)
	}
	defer shutdown()

	reporter.Init(tracer.HookForLogger())

	e := echo.New()
	registerMiddleware(e)

	mainGroup := e.Group("")

	mux, err := gateway.Gateway(ctx)
	if err != nil {
		panic(err)
	}

	mainGroup.Any("/*", echo.WrapHandler(mux))

	if err := e.Start(fmt.Sprintf("%s:%d", *listenHost, *listenPort)); err != nil {
		panic(err)
	}
}

func registerMiddleware(e *echo.Echo) {
	e.Use(echoMiddleware.LoggerWithConfig(echoMiddleware.LoggerConfig{
		Format: "${status} ${method} ${uri}",
		Output: log.With().Str("service", "echo").Logger(),
	}))
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.Gzip())
	e.Use(echoMiddleware.BodyLimit("2M"))
	e.Use(otelecho.Middleware(*appName, otelecho.WithTracerProvider(tracer.Provider)))
	e.Use(tracer.AddTraceIDToResponse)
}
