package repository

import (
	"context"

	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
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

func (r *bibleRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *bibleRepository) GetVerses(
	ctx context.Context,
	query entity.VerseQuery,
) ([]*entity.BibleVerse, error) {
	r.log.Info("get verses started", zap.Any("param", query))

	q := r.conn(ctx).
		Where("translation = ? AND book_id = ? AND chapter = ? AND verse >= ?",
			query.Translation, query.BookID, query.Chapter, query.VerseStart)

	if query.VerseEnd != nil {
		q = q.Where("verse <= ?", *query.VerseEnd)
	}

	var verses []*entity.BibleVerse

	if err := q.Order("verse ASC").Find(&verses).Error; err != nil {
		r.log.Error(
			"get verses failed",
			zap.Error(err),
			zap.Any("param", query),
		)

		return nil, entityerror.ErrBibleRetrievalFailed(err)
	}

	r.log.Info("get verses completed", zap.Any("param", query))

	return verses, nil
}

func (r *bibleRepository) SearchByText(
	ctx context.Context,
	translation string,
	query string,
) ([]*entity.BibleVerse, error) {
	r.log.Info("search by text started", zap.Any("param", query))

	pattern := "%" + query + "%"

	var verses []*entity.BibleVerse

	err := r.conn(ctx).Table("bible_verses").
		Joins("JOIN bible_fts ON bible_fts.verse_id = bible_verses.id").
		Where("bible_fts.translation = ? AND (bible_fts.text LIKE ? OR bible_fts.display_ref LIKE ?)", translation, pattern, pattern).
		Order("bible_verses.book_id ASC, bible_verses.chapter ASC, bible_verses.verse ASC").
		Find(&verses).Error

	if err != nil {
		r.log.Error(
			"search by text failed",
			zap.Error(err),
			zap.Any("param", query),
		)

		return nil, entityerror.ErrBibleRetrievalFailed(err)
	}

	r.log.Info("search by text completed", zap.Any("param", query))

	return verses, nil
}

func (r *bibleRepository) ImportTranslationPack(
	ctx context.Context,
	verses []entity.BibleVerse,
) error {
	r.log.Info("import translation pack started", zap.Any("param", len(verses)))

	db := r.conn(ctx)

	for i := range verses {
		if err := db.Create(&verses[i]).Error; err != nil {
			r.log.Error(
				"import translation pack failed",
				zap.Error(err),
				zap.Any("param", len(verses)),
			)

			return entityerror.ErrBibleImportTranslationPackFailed(err)
		}

		fts := entity.BibleFTS{
			VerseID:     verses[i].ID,
			Translation: verses[i].Translation,
			DisplayRef:  verses[i].BookName,
			Text:        verses[i].Text,
		}

		if err := db.Create(&fts).Error; err != nil {
			r.log.Error(
				"import translation pack failed",
				zap.Error(err),
				zap.Any("param", len(verses)),
			)

			return entityerror.ErrBibleImportTranslationPackFailed(err)
		}
	}

	r.log.Info(
		"import translation pack completed",
		zap.Any("param", len(verses)),
	)

	return nil
}

func (r *bibleRepository) ListAvailableTranslations(
	ctx context.Context,
) ([]string, error) {
	r.log.Info("list available translations started", zap.Any("param", nil))

	var translations []string

	bibleVerseModel := &entity.BibleVerse{}
	if err := r.conn(ctx).Model(bibleVerseModel).Distinct().Pluck("translation", &translations).Error; err != nil {
		r.log.Error(
			"list available translations failed",
			zap.Error(err),
			zap.Any("param", nil),
		)

		return nil, entityerror.ErrBibleRetrievalFailed(err)
	}

	r.log.Info("list available translations completed", zap.Any("param", nil))

	return translations, nil
}
