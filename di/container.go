//go:build wireinject
// +build wireinject

package di

import (
	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/fs"
	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/route"
	pullsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/pull_sync_changes"
	pushsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/push_sync_changes"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
	"github.com/google/wire"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Container struct {
	Config *viper.Viper
	Log    *zap.Logger
	App    *echo.Echo

	TxManager db.TxManager

	FsManager fs.Manager

	RouteConfig *route.RouteConfig

	EtcController        *deliveryhttp.EtcController
	ErrorController      *deliveryhttp.ErrorController
	SyncController       *deliveryhttp.SyncController
	SermonNoteController *deliveryhttp.SermonNoteController

	PullSyncChangesUsecase                *pullsyncchanges.PullSyncChangesUsecase
	PushSyncChangesUsecase                *pushsyncchanges.PushSyncChangesUsecase
	UploadSermonNoteAttachmentFileUsecase *uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecase
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
