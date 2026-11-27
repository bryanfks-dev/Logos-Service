package pullsyncchanges

import (
	"context"
	"errors"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	"go.uber.org/zap"
)

type SyncPullChangesResult struct {
	ServerTime   time.Time
	SermonNotes  []*entity.SermonNote
	Gatherings   []*entity.Gathering
	SermonSeries []*entity.SermonSeries
	Preachers    []*entity.Preacher
	Tags         []*entity.Tag
}

func (u *PullSyncChangesUsecase) Call(
	ctx context.Context,
	since time.Time,
) (*SyncPullChangesResult, error) {
	u.log.Info("pull sync changes started", zap.Time("since", since))

	serverTime := time.Now().UTC()

	notes, err := u.sermonNoteRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error("failed to pull sermon notes", zap.Error(err))

		return nil, err
	}

	gatherings, err := u.gatheringRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error("failed to pull gatherings", zap.Error(err))

		return nil, err
	}

	series, err := u.sermonSeriesRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error("failed to pull sermon series", zap.Error(err))

		return nil, err
	}

	preachers, err := u.preacherRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error("failed to pull preachers", zap.Error(err))

		return nil, err
	}

	tags, err := u.tagRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error("failed to pull tags", zap.Error(err))

		return nil, err
	}

	result := &SyncPullChangesResult{
		ServerTime:   serverTime,
		SermonNotes:  notes,
		Gatherings:   gatherings,
		SermonSeries: series,
		Preachers:    preachers,
		Tags:         tags,
	}

	u.log.Info(
		"pull sync changes completed",
		zap.Time("server_time", serverTime),
		zap.Int("sermon_notes_count", len(notes)),
		zap.Int("gatherings_count", len(gatherings)),
		zap.Int("sermon_series_count", len(series)),
		zap.Int("preachers_count", len(preachers)),
		zap.Int("tags_count", len(tags)),
	)

	return result, nil
}
