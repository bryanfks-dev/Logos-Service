package di

import (
	"github.com/bryanfks-dev/selah/internal/delivery"
	"github.com/bryanfks-dev/selah/internal/delivery/route"
	"github.com/bryanfks-dev/selah/pkg/config"
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
	delivery.NewErrorController,
)

var RouteSet = wire.NewSet(
	route.NewRouteConfig,
)
