package repository

import (
	"context"

	dbcorrelation "github.com/bryanfks-dev/selah/infrastructure/db/correlation"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagRepository interface {
	Create(ctx context.Context, tag *entity.Tag) error
	Update(ctx context.Context, tag *entity.Tag) error
	FindAll(ctx context.Context) ([]entity.Tag, error)
	Delete(ctx context.Context, id string) error
}

type tagRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewTagRepository(db *gorm.DB, log *zap.Logger) TagRepository {
	return &tagRepository{
		db:  db,
		log: log,
	}
}

func (r *tagRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := dbcorrelation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *tagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	r.log.Info("create started", zap.Any("param", tag))

	err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(tag).Error

	if err != nil {
		r.log.Error("create failed", zap.Error(err), zap.Any("param", tag))

		return entityerror.ErrTagCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", tag))

	return nil
}

func (r *tagRepository) Update(ctx context.Context, tag *entity.Tag) error {
	r.log.Info("update started", zap.Any("param", tag))

	if err := r.getDB(ctx).Save(tag).Error; err != nil {
		r.log.Error("update failed", zap.Error(err), zap.Any("param", tag))

		return entityerror.ErrTagUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", tag))

	return nil
}

func (r *tagRepository) FindAll(ctx context.Context) ([]entity.Tag, error) {
	r.log.Info("find all started", zap.Any("param", nil))

	type tagWithCount struct {
		entity.Tag
		UsageCount int `gorm:"column:usage_count"`
	}

	var results []tagWithCount

	err := r.getDB(ctx).Table("tags").
		Select("tags.*, count(note_tags.note_id) as usage_count").
		Joins("left join note_tags on note_tags.tag_id = tags.id").
		Group("tags.id").
		Find(&results).Error

	if err != nil {
		r.log.Error("find all failed", zap.Error(err), zap.Any("param", nil))

		return nil, entityerror.ErrTagFindAllFailed(err)
	}

	tags := make([]entity.Tag, len(results))

	for i, res := range results {
		tags[i] = res.Tag
		tags[i].UsageCount = res.UsageCount
	}

	r.log.Info("find all completed", zap.Any("param", nil))

	return tags, nil
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	if err := r.getDB(ctx).Where("id = ?", id).Delete(&entity.Tag{}).Error; err != nil {
		r.log.Error("delete failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrTagDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
