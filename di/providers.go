package di

import (
	"github.com/bryanfks-dev/selah/infrastructure/broker"
	"github.com/bryanfks-dev/selah/infrastructure/db"
	"github.com/bryanfks-dev/selah/internal/delivery"
	"github.com/bryanfks-dev/selah/internal/delivery/route"
	"github.com/bryanfks-dev/selah/internal/repository"
	"github.com/bryanfks-dev/selah/internal/usecase"
	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/google/wire"
)

var AppSet = wire.NewSet(
	config.NewViper,
	config.NewLogger,
	config.NewServer,
	config.NewDatabase,
)

var BrokerSet = wire.NewSet(
	config.NewKafkaConfig,
	broker.NewKafkaConsumer,
	broker.NewKafkaProducer,
)

var InfrastructureSet = wire.NewSet(
	db.NewTxManager,
)

var RepositorySet = wire.NewSet(
	repository.NewGatheringRepository,
	repository.NewSermonSeriesRepository,
	repository.NewPreacherRepository,
	repository.NewTagRepository,
	repository.NewSermonNoteRepository,
	repository.NewBibleRepository,
	repository.NewSyncOutboxRepository,
)

var UsecaseSet = wire.NewSet(
	usecase.NewSyncUsecase,
)

var DeliverySet = wire.NewSet(
	delivery.NewEtcController,
	delivery.NewErrorController,
	delivery.NewSermonNoteController,
)

var RouteSet = wire.NewSet(
	route.NewRouteConfig,
)
