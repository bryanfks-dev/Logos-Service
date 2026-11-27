package di

import (
	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/route"
	"github.com/bryanfks-dev/logos/internal/repository/bible"
	"github.com/bryanfks-dev/logos/internal/repository/gathering"
	"github.com/bryanfks-dev/logos/internal/repository/preacher"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	sermonseries "github.com/bryanfks-dev/logos/internal/repository/sermon_series"
	syncoutbox "github.com/bryanfks-dev/logos/internal/repository/sync_outbox"
	"github.com/bryanfks-dev/logos/internal/repository/tag"
	downloadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/download_sermon_note_attachment_file"
	pullsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/pull_sync_changes"
	pushsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/push_sync_changes"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
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
	fs.NewManager,
)

var RepositorySet = wire.NewSet(
	gathering.NewGatheringRepository,
	sermonseries.NewSermonSeriesRepository,
	preacher.NewPreacherRepository,
	tag.NewTagRepository,
	sermonnote.NewSermonNoteRepository,
	bible.NewBibleRepository,
	syncoutbox.NewSyncOutboxRepository,
)

var UsecaseSet = wire.NewSet(
	wire.Struct(new(pullsyncchanges.PullSyncChangesUsecaseOption), "*"),
	pullsyncchanges.NewPullSyncChangesUsecase,
	wire.Struct(new(pushsyncchanges.PushSyncChangesUsecaseOption), "*"),
	pushsyncchanges.NewPushSyncChangesUsecase,
	wire.Struct(
		new(
			uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecaseOption,
		),
		"*",
	),
	uploadsermonnoteattachmentfile.NewUploadSermonNoteAttachmentFileUsecase,
	downloadsermonnoteattachmentfile.NewDownloadSermonNoteAttachmentFileUsecase,
)

var DeliverySet = wire.NewSet(
	http.NewEtcController,
	http.NewErrorController,
	http.NewSyncController,
	http.NewSermonNoteController,
)

var RouteSet = wire.NewSet(
	wire.Struct(new(route.RouteConfigOption), "*"),
	route.NewRouteConfig,
)
