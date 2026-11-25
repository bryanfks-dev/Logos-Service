package usecase

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/repository"
	"go.uber.org/zap"
)

type PullSyncChangesUsecase struct {
	sermonNoteRepository   repository.SermonNoteRepository
	gatheringRepository    repository.GatheringRepository
	sermonSeriesRepository repository.SermonSeriesRepository
	preacherRepository     repository.PreacherRepository
	tagRepository          repository.TagRepository
	log                    *zap.Logger
}

type PullSyncChangesUsecaseOption struct {
	SermonNoteRepository   repository.SermonNoteRepository
	GatheringRepository    repository.GatheringRepository
	SermonSeriesRepository repository.SermonSeriesRepository
	PreacherRepository     repository.PreacherRepository
	TagRepository          repository.TagRepository
	Log                    *zap.Logger
}

func NewPullSyncChangesUsecase(
	opt *PullSyncChangesUsecaseOption,
) *PullSyncChangesUsecase {
	return &PullSyncChangesUsecase{
		sermonNoteRepository:   opt.SermonNoteRepository,
		gatheringRepository:    opt.GatheringRepository,
		sermonSeriesRepository: opt.SermonSeriesRepository,
		preacherRepository:     opt.PreacherRepository,
		tagRepository:          opt.TagRepository,
		log:                    opt.Log,
	}
}

func (u *PullSyncChangesUsecase) Call(
	ctx context.Context,
	since time.Time,
) (*model.SyncPullChangesResult, error) {
	serverTime := time.Now().UTC()

	notes, err := u.sermonNoteRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		u.log.Error("failed to pull sermon notes", zap.Error(err))
		return nil, err
	}

	gatherings, err := u.gatheringRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		u.log.Error("failed to pull gatherings", zap.Error(err))
		return nil, err
	}

	series, err := u.sermonSeriesRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		u.log.Error("failed to pull sermon series", zap.Error(err))
		return nil, err
	}

	preachers, err := u.preacherRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		u.log.Error("failed to pull preachers", zap.Error(err))
		return nil, err
	}

	tags, err := u.tagRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		u.log.Error("failed to pull tags", zap.Error(err))
		return nil, err
	}

	result := &model.SyncPullChangesResult{
		ServerTime:   serverTime,
		SermonNotes:  notes,
		Gatherings:   gatherings,
		SermonSeries: series,
		Preachers:    preachers,
		Tags:         tags,
	}

	return result, nil
}
