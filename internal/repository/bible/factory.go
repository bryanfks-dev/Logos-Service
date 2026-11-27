package bible

import (
	"context"

	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BibleRepository interface {
	GetVerses(
		ctx context.Context,
		query entity.VerseQuery,
	) ([]*entity.BibleVerse, error)
	SearchByText(
		ctx context.Context,
		translation string,
		query string,
	) ([]*entity.BibleVerse, error)
	ImportTranslationPack(ctx context.Context, verses []entity.BibleVerse) error
	ListAvailableTranslations(ctx context.Context) ([]string, error)
}

type bibleRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewBibleRepository(db *gorm.DB, log *zap.Logger) BibleRepository {
	return &bibleRepository{
		db:  db,
		log: log,
	}
}
