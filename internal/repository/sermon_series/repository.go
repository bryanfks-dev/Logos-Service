package sermonseries

import (
	"context"
	"errors"
	"time"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *sermonSeriesRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.GetTx(ctx); tx != nil {
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
		return entityerror.ErrSermonSeriesCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", series))

	return nil
}

func (r *sermonSeriesRepository) CreateBatch(
	ctx context.Context,
	series []*entity.SermonSeries,
) error {
	r.log.Info("create batch started", zap.Int("count", len(series)))

	if len(series) == 0 {
		r.log.Debug("create batch skipped due to empty slice")

		return nil
	}

	if err := r.conn(ctx).CreateInBatches(series, 500).Error; err != nil {
		return entityerror.ErrSermonSeriesCreationFailed(err)
	}

	r.log.Info("create batch completed", zap.Int("count", len(series)))

	return nil
}

func (r *sermonSeriesRepository) Update(
	ctx context.Context,
	series *entity.SermonSeries,
) error {
	r.log.Info("update started", zap.Any("param", series))

	if err := r.conn(ctx).Save(series).Error; err != nil {
		return entityerror.ErrSermonSeriesUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", series))

	return nil
}

func (r *sermonSeriesRepository) UpdateBatch(
	ctx context.Context,
	series []*entity.SermonSeries,
) error {
	r.log.Info("update batch started", zap.Int("count", len(series)))

	if len(series) == 0 {
		r.log.Debug("update batch skipped due to empty slice")

		return nil
	}

	onConflict := clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}

	if err := r.conn(ctx).Clauses(onConflict).CreateInBatches(series, 500).Error; err != nil {
		return entityerror.ErrSermonSeriesUpdateFailed(err)
	}

	r.log.Info("update batch completed", zap.Int("count", len(series)))

	return nil
}

func (r *sermonSeriesRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.SermonSeries, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var item entity.SermonSeries

	if err := r.conn(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonSeriesNotFound(id)
		}

		return nil, entityerror.ErrSermonSeriesRetrievalFailed(err)
	}

	r.log.Info("find by id completed", zap.Any("param", id))

	return &item, nil
}

func (r *sermonSeriesRepository) FindByIDs(
	ctx context.Context,
	ids []string,
) ([]*entity.SermonSeries, error) {
	r.log.Info("find by ids started", zap.Strings("ids", ids))

	if len(ids) == 0 {
		r.log.Debug("find by ids skipped due to empty ids")

		return make([]*entity.SermonSeries, 0), nil
	}

	var items []*entity.SermonSeries

	query := db.WithRowLock(
		ctx,
		r.conn(ctx).Unscoped().Where("id IN ?", ids),
	)
	if err := query.Find(&items).Error; err != nil {
		return nil, entityerror.ErrSermonSeriesRetrievalFailed(err)
	}

	r.log.Info("find by ids completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *sermonSeriesRepository) FindAll(
	ctx context.Context,
) ([]*entity.SermonSeries, error) {
	r.log.Info("find all started", zap.Any("param", nil))

	var items []*entity.SermonSeries

	if err := r.conn(ctx).Find(&items).Error; err != nil {
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
		return nil, entityerror.ErrSermonSeriesRetrievalFailed(err)
	}

	r.log.Info("find updated since completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *sermonSeriesRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	seriesModel := &entity.SermonSeries{}
	if err := r.conn(ctx).Where("id = ?", id).Delete(seriesModel).Error; err != nil {
		return entityerror.ErrSermonSeriesDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
