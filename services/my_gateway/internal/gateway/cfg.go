package gateway

import zfg "github.com/chaindead/zerocfg"

var (
	connectionsGroup            = zfg.NewGroup("connection")
	connectionStringSomeService = zfg.Str("some_service", "0.0.0.0:11000", "CONNECTION_SOMESERVICE", zfg.Group(connectionsGroup))
	connectionStringAuthService = zfg.Str("auth_service", "0.0.0.0:12000", "CONNECTION_AUTHSERVICE", zfg.Group(connectionsGroup))

	// example declaring connection strings config
	//connectionStringYetAnotherService = zfg.Str("yet_another_service", "0.0.0.0:11000", "CONNECTION_YETANOTHERSERVICE", zfg.Group(connectionsGroup))
)
