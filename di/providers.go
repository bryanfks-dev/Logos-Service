package di

import (
	"github.com/bryanfks-dev/logos/db"
	"github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/route"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/internal/usecase"
	"github.com/bryanfks-dev/logos/pkg/config"
	"github.com/google/wire"
)

var AppSet = wire.NewSet(
	config.NewViper,
	config.NewLogger,
	config.NewServer,
	config.NewDatabase,
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
	wire.Struct(new(usecase.SyncChangesUsecaseOption), "*"),
	usecase.NewSyncChangesUsecase,
)

var DeliverySet = wire.NewSet(
	http.NewEtcController,
	http.NewErrorController,
	http.NewSyncController,
)

var RouteSet = wire.NewSet(
	wire.Struct(new(route.RouteConfigOption), "*"),
	route.NewRouteConfig,
)
