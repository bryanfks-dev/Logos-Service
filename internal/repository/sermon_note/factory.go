package sermonnote

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SermonNoteRepository interface {
	Create(ctx context.Context, note *entity.SermonNote) error
	CreateBatch(ctx context.Context, notes []*entity.SermonNote) error
	Update(ctx context.Context, note *entity.SermonNote) error
	UpdateBatch(ctx context.Context, notes []*entity.SermonNote) error
	FindByID(ctx context.Context, id string) (*entity.SermonNote, error)
	FindByIDUnscoped(ctx context.Context, id string) (*entity.SermonNote, error)
	FindByIDsUnscoped(
		ctx context.Context,
		ids []string,
	) ([]*entity.SermonNote, error)
	FindAll(
		ctx context.Context,
		filter entity.SermonNoteFilter,
	) ([]*entity.SermonNote, error)
	Search(ctx context.Context, query string) ([]*entity.SermonNote, error)
	FindByScripture(
		ctx context.Context,
		query entity.ScriptureQuery,
	) ([]*entity.SermonNote, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.SermonNote, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	FindAttachmentByID(
		ctx context.Context,
		id string,
	) (*entity.SermonNoteAttachment, error)
	UpdateAttachment(
		ctx context.Context,
		attachment *entity.SermonNoteAttachment,
	) error
}

type sermonNoteRepository struct {
	db        *gorm.DB
	fsManager fs.Manager
	log       *zap.Logger
}

func NewSermonNoteRepository(
	db *gorm.DB,
	fsManager fs.Manager,
	log *zap.Logger,
) SermonNoteRepository {
	return &sermonNoteRepository{
		db:        db,
		fsManager: fsManager,
		log:       log,
	}
}
