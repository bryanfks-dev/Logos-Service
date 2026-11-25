package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"github.com/bryanfks-dev/logos/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PreacherRepository interface {
	Create(ctx context.Context, preacher *entity.Preacher) error
	Update(ctx context.Context, preacher *entity.Preacher) error
	FindByID(ctx context.Context, id string) (*entity.Preacher, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Preacher, error)
	FindAll(
		ctx context.Context,
		filter entity.PreacherFilter,
	) ([]entity.Preacher, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.Preacher, error)
	Delete(ctx context.Context, id string) error
}

type preacherRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewPreacherRepository(db *gorm.DB, log *zap.Logger) PreacherRepository {
	return &preacherRepository{
		db:  db,
		log: log,
	}
}

func (r *preacherRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *preacherRepository) Create(
	ctx context.Context,
	preacher *entity.Preacher,
) error {
	r.log.Info("create started", zap.Any("param", preacher))

	if err := r.conn(ctx).Create(preacher).Error; err != nil {
		r.log.Error("create failed", zap.Error(err), zap.Any("param", preacher))

		return entityerror.ErrPreacherCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", preacher))

	return nil
}

func (r *preacherRepository) Update(
	ctx context.Context,
	preacher *entity.Preacher,
) error {
	r.log.Info("update started", zap.Any("param", preacher))

	if err := r.conn(ctx).Save(preacher).Error; err != nil {
		r.log.Error("update failed", zap.Error(err), zap.Any("param", preacher))

		return entityerror.ErrPreacherUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", preacher))

	return nil
}

func (r *preacherRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.Preacher, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var item entity.Preacher

	if err := r.conn(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		r.log.Error("find by id failed", zap.Error(err), zap.Any("param", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrPreacherNotFound(id)
		}

		return nil, entityerror.ErrPreacherRetrievalFailed(err)
	}

	r.log.Info("find by id completed", zap.Any("param", id))

	return &item, nil
}

func (r *preacherRepository) FindByIDs(
	ctx context.Context,
	ids []string,
) ([]*entity.Preacher, error) {
	if len(ids) == 0 {
		return make([]*entity.Preacher, 0), nil
	}

	var items []*entity.Preacher

	query := db.WithRowLock(
		ctx,
		r.conn(ctx).Unscoped().Where("id IN ?", ids),
	)
	if err := query.Find(&items).Error; err != nil {
		r.log.Error(
			"find by ids failed",
			zap.Error(err),
			zap.Strings("ids", ids),
		)

		return nil, entityerror.ErrPreacherRetrievalFailed(err)
	}

	return items, nil
}

func (r *preacherRepository) FindAll(
	ctx context.Context,
	filter entity.PreacherFilter,
) ([]entity.Preacher, error) {
	r.log.Info("find all started", zap.Any("param", filter))

	query := r.conn(ctx)

	if filter.IncludeDeleted {
		query = query.Unscoped()
	}

	if filter.NamePrefix != "" {
		query = query.Where("name LIKE ?", filter.NamePrefix+"%")
	}

	var items []entity.Preacher

	if err := query.Find(&items).Error; err != nil {
		r.log.Error("find all failed", zap.Error(err), zap.Any("param", filter))

		return nil, entityerror.ErrPreacherRetrievalFailed(err)
	}

	r.log.Info("find all completed", zap.Any("param", filter))

	return items, nil
}

func (r *preacherRepository) FindUpdatedSince(
	ctx context.Context,
	since time.Time,
) ([]*entity.Preacher, error) {
	r.log.Info("find updated since started", zap.Time("since", since))

	var items []*entity.Preacher

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

		return nil, entityerror.ErrPreacherRetrievalFailed(err)
	}

	items = util.NonNilSlice(items)

	r.log.Info("find updated since completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *preacherRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	preacherModel := &entity.Preacher{}
	if err := r.conn(ctx).Where("id = ?", id).Delete(preacherModel).Error; err != nil {
		r.log.Error("delete failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrPreacherDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
