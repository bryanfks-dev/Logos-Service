//go:build wireinject
// +build wireinject

package di

import (
	"github.com/bryanfks-dev/selah/internal/delivery"
	"github.com/bryanfks-dev/selah/internal/delivery/route"
	"github.com/google/wire"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Container struct {
	Config *viper.Viper
	Log    *zap.Logger
	App    *echo.Echo

	RouteConfig *route.RouteConfig

	EtcController   *delivery.EtcController
	ErrorController *delivery.ErrorController
}

func NewContainer() (*Container, error) {
	wire.Build(
		AppSet,
		DeliverySet,
		RouteSet,
		wire.Struct(new(Container), "*"),
	)

	return nil, nil
}
