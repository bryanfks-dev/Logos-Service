package uploadsermonnoteattachmentfile

import (
	"github.com/bryanfks-dev/logos/infra/fs"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type UploadSermonNoteAttachmentFileUsecase struct {
	sermonNoteRepository sermonnote.SermonNoteRepository
	fsManager            fs.Manager
	appPath              string
	log                  *zap.Logger
}

type UploadSermonNoteAttachmentFileUsecaseOption struct {
	SermonNoteRepository sermonnote.SermonNoteRepository
	FsManager            fs.Manager
	Config               *viper.Viper
	Log                  *zap.Logger
}

func NewUploadSermonNoteAttachmentFileUsecase(
	opt *UploadSermonNoteAttachmentFileUsecaseOption,
) *UploadSermonNoteAttachmentFileUsecase {
	return &UploadSermonNoteAttachmentFileUsecase{
		sermonNoteRepository: opt.SermonNoteRepository,
		fsManager:            opt.FsManager,
		appPath:              opt.Config.GetString("app.path"),
		log:                  opt.Log,
	}
}
