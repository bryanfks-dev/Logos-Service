//go:build wireinject
// +build wireinject

package di

import (
	"github.com/bryanfks-dev/selah/infrastructure/broker"
	"github.com/bryanfks-dev/selah/infrastructure/db"
	"github.com/bryanfks-dev/selah/internal/delivery"
	"github.com/bryanfks-dev/selah/internal/delivery/route"
	"github.com/bryanfks-dev/selah/internal/usecase"
	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/google/wire"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Container struct {
	Config *viper.Viper
	Log    *zap.Logger
	App    *echo.Echo

	KafkaConfig    *config.KafkaConfig
	BrokerConsumer broker.Consumer
	BrokerProducer broker.Producer

	TxManager *db.TxManager

	RouteConfig *route.RouteConfig

	EtcController        *delivery.EtcController
	ErrorController      *delivery.ErrorController
	SermonNoteController *delivery.SermonNoteController

	SyncUsecase usecase.SyncUsecase
}

func NewContainer() (*Container, error) {
	wire.Build(
		AppSet,
		BrokerSet,
		InfrastructureSet,
		RepositorySet,
		UsecaseSet,
		DeliverySet,
		RouteSet,
		wire.Struct(new(Container), "*"),
	)

	return nil, nil
}
