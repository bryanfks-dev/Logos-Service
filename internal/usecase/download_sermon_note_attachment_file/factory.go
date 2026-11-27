package downloadsermonnoteattachmentfile

import (
	"github.com/bryanfks-dev/logos/infra/fs"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	"go.uber.org/zap"
)

type DownloadSermonNoteAttachmentFileUsecase struct {
	sermonNoteRepository sermonnote.SermonNoteRepository
	fsManager            fs.Manager
	log                  *zap.Logger
}

func NewDownloadSermonNoteAttachmentFileUsecase(
	sermonNoteRepository sermonnote.SermonNoteRepository,
	fsManager fs.Manager,
	log *zap.Logger,
) *DownloadSermonNoteAttachmentFileUsecase {
	return &DownloadSermonNoteAttachmentFileUsecase{
		sermonNoteRepository: sermonNoteRepository,
		fsManager:            fsManager,
		log:                  log,
	}
}
