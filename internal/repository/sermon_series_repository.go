package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bryanfks-dev/logos/db/correlation"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"github.com/bryanfks-dev/logos/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SermonSeriesRepository interface {
	Create(ctx context.Context, series *entity.SermonSeries) error
	Update(ctx context.Context, series *entity.SermonSeries) error
	FindByID(ctx context.Context, id string) (*entity.SermonSeries, error)
	FindAll(ctx context.Context) ([]entity.SermonSeries, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.SermonSeries, error)
	Delete(ctx context.Context, id string) error
}

type sermonSeriesRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSermonSeriesRepository(
	db *gorm.DB,
	log *zap.Logger,
) SermonSeriesRepository {
	return &sermonSeriesRepository{
		db:  db,
		log: log,
	}
}

func (r *sermonSeriesRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.Tx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *sermonSeriesRepository) Create(
	ctx context.Context,
	series *entity.SermonSeries,
) error {
	r.log.Info("create started", zap.Any("param", series))

	if err := r.conn(ctx).Create(series).Error; err != nil {
		r.log.Error("create failed", zap.Error(err), zap.Any("param", series))

		return entityerror.ErrSermonSeriesCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", series))

	return nil
}

func (r *sermonSeriesRepository) Update(
	ctx context.Context,
	series *entity.SermonSeries,
) error {
	r.log.Info("update started", zap.Any("param", series))

	if err := r.conn(ctx).Save(series).Error; err != nil {
		r.log.Error("update failed", zap.Error(err), zap.Any("param", series))

		return entityerror.ErrSermonSeriesUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", series))

	return nil
}

func (r *sermonSeriesRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.SermonSeries, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var item entity.SermonSeries

	if err := r.conn(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		r.log.Error("find by id failed", zap.Error(err), zap.Any("param", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonSeriesNotFound(id)
		}

		return nil, entityerror.ErrSermonSeriesRetrievalFailed(err)
	}

	r.log.Info("find by id completed", zap.Any("param", id))

	return &item, nil
}

func (r *sermonSeriesRepository) FindAll(
	ctx context.Context,
) ([]entity.SermonSeries, error) {
	r.log.Info("find all started", zap.Any("param", nil))

	var items []entity.SermonSeries

	if err := r.conn(ctx).Find(&items).Error; err != nil {
		r.log.Error("find all failed", zap.Error(err), zap.Any("param", nil))

		return nil, entityerror.ErrSermonSeriesRetrievalFailed(err)
	}

	r.log.Info("find all completed", zap.Any("param", nil))

	return items, nil
}

func (r *sermonSeriesRepository) FindUpdatedSince(
	ctx context.Context,
	since time.Time,
) ([]*entity.SermonSeries, error) {
	r.log.Info("find updated since started", zap.Time("since", since))

	var items []*entity.SermonSeries

	query := r.conn(ctx).Unscoped()
	if !since.IsZero() {
		query = query.Where("updated_at > ?", since)
	}

	err := query.Order("updated_at ASC").Find(&items).Error
	if err != nil {
		r.log.Error(
			"find updated since failed",
			zap.Error(err),
			zap.Time("since", since),
		)

		return nil, entityerror.ErrSermonSeriesRetrievalFailed(err)
	}

	items = util.NonNilSlice(items)

	r.log.Info("find updated since completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *sermonSeriesRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	if err := r.conn(ctx).Where("id = ?", id).Delete(&entity.SermonSeries{}).Error; err != nil {
		r.log.Error("delete failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrSermonSeriesDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
