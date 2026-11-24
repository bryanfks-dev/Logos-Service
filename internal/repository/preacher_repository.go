package repository

import (
	"context"
	"errors"

	dbcorrelation "github.com/bryanfks-dev/selah/infrastructure/db/correlation"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PreacherRepository interface {
	Create(ctx context.Context, preacher *entity.Preacher) error
	Update(ctx context.Context, preacher *entity.Preacher) error
	FindByID(ctx context.Context, id string) (*entity.Preacher, error)
	FindAll(
		ctx context.Context,
		filter entity.PreacherFilter,
	) ([]entity.Preacher, error)
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

func (r *preacherRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := dbcorrelation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *preacherRepository) Create(
	ctx context.Context,
	preacher *entity.Preacher,
) error {
	r.log.Info("create started", zap.Any("param", preacher))

	if err := r.getDB(ctx).Create(preacher).Error; err != nil {
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

	if err := r.getDB(ctx).Save(preacher).Error; err != nil {
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

	if err := r.getDB(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		r.log.Error("find by id failed", zap.Error(err), zap.Any("param", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrPreacherNotFound(id)
		}

		return nil, entityerror.ErrPreacherFindByIDFailed(err)
	}

	r.log.Info("find by id completed", zap.Any("param", id))

	return &item, nil
}

func (r *preacherRepository) FindAll(
	ctx context.Context,
	filter entity.PreacherFilter,
) ([]entity.Preacher, error) {
	r.log.Info("find all started", zap.Any("param", filter))

	query := r.getDB(ctx)

	if filter.IncludeDeleted {
		query = query.Unscoped()
	}

	if filter.NamePrefix != "" {
		query = query.Where("name LIKE ?", filter.NamePrefix+"%")
	}

	var items []entity.Preacher

	if err := query.Find(&items).Error; err != nil {
		r.log.Error("find all failed", zap.Error(err), zap.Any("param", filter))

		return nil, entityerror.ErrPreacherFindAllFailed(err)
	}

	r.log.Info("find all completed", zap.Any("param", filter))

	return items, nil
}

func (r *preacherRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	if err := r.getDB(ctx).Where("id = ?", id).Delete(&entity.Preacher{}).Error; err != nil {
		r.log.Error("delete failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrPreacherDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
