//go:build wireinject
// +build wireinject

package di

import (
	"github.com/bryanfks-dev/logos/db"
	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/route"
	"github.com/bryanfks-dev/logos/internal/usecase"
	"github.com/google/wire"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Container struct {
	Config *viper.Viper
	Log    *zap.Logger
	App    *echo.Echo

	TxManager *db.TxManager

	RouteConfig *route.RouteConfig

	EtcController   *deliveryhttp.EtcController
	ErrorController *deliveryhttp.ErrorController
	SyncHandler     *deliveryhttp.SyncHandler

	SyncUsecase *usecase.SyncUsecase
}

func NewContainer() (*Container, error) {
	wire.Build(
		AppSet,
		InfrastructureSet,
		RepositorySet,
		UsecaseSet,
		DeliverySet,
		RouteSet,
		wire.Struct(new(Container), "*"),
	)

	return nil, nil
}
