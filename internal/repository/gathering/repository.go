package gathering

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

func (r *gatheringRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *gatheringRepository) Create(
	ctx context.Context,
	gathering *entity.Gathering,
) error {
	r.log.Info("create started", zap.Any("param", gathering))

	if err := r.conn(ctx).Create(gathering).Error; err != nil {
		return entityerror.ErrGatheringCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", gathering))

	return nil
}

func (r *gatheringRepository) CreateBatch(
	ctx context.Context,
	gatherings []*entity.Gathering,
) error {
	r.log.Info("create batch started", zap.Int("count", len(gatherings)))

	if len(gatherings) == 0 {
		r.log.Debug("create batch skipped due to empty slice")

		return nil
	}

	if err := r.conn(ctx).CreateInBatches(gatherings, 500).Error; err != nil {
		return entityerror.ErrGatheringCreationFailed(err)
	}

	r.log.Info("create batch completed", zap.Int("count", len(gatherings)))

	return nil
}

func (r *gatheringRepository) Update(
	ctx context.Context,
	gathering *entity.Gathering,
) error {
	r.log.Info("update started", zap.Any("param", gathering))

	if err := r.conn(ctx).Save(gathering).Error; err != nil {
		return entityerror.ErrGatheringUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", gathering))

	return nil
}

func (r *gatheringRepository) UpdateBatch(
	ctx context.Context,
	gatherings []*entity.Gathering,
) error {
	r.log.Info("update batch started", zap.Int("count", len(gatherings)))

	if len(gatherings) == 0 {
		r.log.Debug("update batch skipped due to empty slice")

		return nil
	}

	onConflict := clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}

	if err := r.conn(ctx).Clauses(onConflict).CreateInBatches(gatherings, 500).Error; err != nil {
		return entityerror.ErrGatheringUpdateFailed(err)
	}

	r.log.Info("update batch completed", zap.Int("count", len(gatherings)))

	return nil
}

func (r *gatheringRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.Gathering, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var item entity.Gathering

	if err := r.conn(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrGatheringNotFound(id)
		}

		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	r.log.Info("find by id completed", zap.Any("param", id))

	return &item, nil
}

func (r *gatheringRepository) FindByIDs(
	ctx context.Context,
	ids []string,
) ([]*entity.Gathering, error) {
	r.log.Info("find by ids started", zap.Strings("ids", ids))

	if len(ids) == 0 {
		r.log.Debug("find by ids skipped due to empty ids")

		return make([]*entity.Gathering, 0), nil
	}

	var items []*entity.Gathering

	query := db.WithRowLock(
		ctx,
		r.conn(ctx).Unscoped().Where("id IN ?", ids),
	)
	if err := query.Find(&items).Error; err != nil {
		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	r.log.Info("find by ids completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *gatheringRepository) FindAll(
	ctx context.Context,
) ([]*entity.Gathering, error) {
	r.log.Info("find all started", zap.Any("param", nil))

	var items []*entity.Gathering

	if err := r.conn(ctx).Order("sort_order ASC").Find(&items).Error; err != nil {
		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	r.log.Info("find all completed", zap.Any("param", nil))

	return items, nil
}

func (r *gatheringRepository) FindUpdatedSince(
	ctx context.Context,
	since time.Time,
) ([]*entity.Gathering, error) {
	r.log.Info("find updated since started", zap.Time("since", since))

	var items []*entity.Gathering

	query := r.conn(ctx).Unscoped()
	if !since.IsZero() {
		query = query.Where("updated_at > ?", since)
	}

	err := query.Order("updated_at ASC").Find(&items).Error
	if err != nil {
		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	r.log.Info("find updated since completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *gatheringRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	gatheringModel := &entity.Gathering{}
	if err := r.conn(ctx).Where("id = ?", id).Delete(gatheringModel).Error; err != nil {
		return entityerror.ErrGatheringDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
