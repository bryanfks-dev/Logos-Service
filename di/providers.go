package di

import (
	"github.com/bryanfks-dev/internal/delivery"
	"github.com/bryanfks-dev/internal/delivery/route"
	"github.com/bryanfks-dev/pkg/config"
	"github.com/google/wire"
)

var AppSet = wire.NewSet(
	config.NewViper,
	config.NewLogger,
	config.NewServer,
	config.NewDatabase,
)

var RepositorySet = wire.NewSet()

var UsecaseSet = wire.NewSet()

var DeliverySet = wire.NewSet(
	delivery.NewEtcController,
)

var RouteSet = wire.NewSet(
	route.NewRouteConfig,
)
