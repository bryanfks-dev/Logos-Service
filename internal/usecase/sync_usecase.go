package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/bryanfks-dev/selah/db"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/repository"
	"github.com/bryanfks-dev/selah/pkg/entity"
	apperror "github.com/bryanfks-dev/selah/pkg/error"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"github.com/bryanfks-dev/selah/pkg/util"
	"go.uber.org/zap"
)

type SyncUsecase struct {
	txManager              *db.TxManager
	sermonNoteRepository   repository.SermonNoteRepository
	gatheringRepository    repository.GatheringRepository
	sermonSeriesRepository repository.SermonSeriesRepository
	preacherRepository     repository.PreacherRepository
	tagRepository          repository.TagRepository
	syncOutboxRepository   repository.SyncOutboxRepository
	log                    *zap.Logger
}

type SyncUsecaseOption struct {
	TxManager              *db.TxManager
	SermonNoteRepository   repository.SermonNoteRepository
	GatheringRepository    repository.GatheringRepository
	SermonSeriesRepository repository.SermonSeriesRepository
	PreacherRepository     repository.PreacherRepository
	TagRepository          repository.TagRepository
	SyncOutboxRepository   repository.SyncOutboxRepository
	Log                    *zap.Logger
}

func NewSyncUsecase(opt *SyncUsecaseOption) *SyncUsecase {
	return &SyncUsecase{
		txManager:              opt.TxManager,
		sermonNoteRepository:   opt.SermonNoteRepository,
		gatheringRepository:    opt.GatheringRepository,
		sermonSeriesRepository: opt.SermonSeriesRepository,
		preacherRepository:     opt.PreacherRepository,
		tagRepository:          opt.TagRepository,
		syncOutboxRepository:   opt.SyncOutboxRepository,
		log:                    opt.Log,
	}
}

func (u *SyncUsecase) Pull(
	ctx context.Context,
	since time.Time,
) (*model.SyncPullResult, error) {
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

	notes = util.NonNilSlice(notes)
	gatherings = util.NonNilSlice(gatherings)
	series = util.NonNilSlice(series)
	preachers = util.NonNilSlice(preachers)
	tags = util.NonNilSlice(tags)

	result := &model.SyncPullResult{
		ServerTime:   serverTime,
		Notes:        notes,
		SermonNotes:  notes,
		Gatherings:   gatherings,
		Series:       series,
		SermonSeries: series,
		Preachers:    preachers,
		Tags:         tags,
	}

	return result, nil
}

func (u *SyncUsecase) Push(
	ctx context.Context,
	batch *model.SyncBatch,
) (*model.SyncPushResult, error) {
	if batch == nil {
		batch = &model.SyncBatch{}
	}

	syncedAt := time.Now().UTC()

	var (
		missingContentIDs []string
		conflicts         []string
	)

	err := u.txManager.Do(ctx, func(txCtx context.Context) error {
		var txErr error
		missingContentIDs, conflicts, txErr = u.applyPushBatch(txCtx, batch)

		return txErr
	})
	if err != nil {
		u.log.Error("failed to push sync batch", zap.Error(err))
		return nil, err
	}

	result := &model.SyncPushResult{
		SyncedAt:          syncedAt,
		MissingContentIDs: missingContentIDs,
		Conflicts:         conflicts,
	}

	return result, nil
}

func (u *SyncUsecase) applyPushBatch(
	ctx context.Context,
	batch *model.SyncBatch,
) ([]string, []string, error) {
	if err := u.upsertGatherings(ctx, batch.Gatherings); err != nil {
		return nil, nil, err
	}

	series := util.CoalesceSlice(batch.Series, batch.SermonSeries)
	if err := u.upsertSermonSeries(ctx, series); err != nil {
		return nil, nil, err
	}

	if err := u.upsertPreachers(ctx, batch.Preachers); err != nil {
		return nil, nil, err
	}

	if err := u.upsertTags(ctx, batch.Tags); err != nil {
		return nil, nil, err
	}

	notes := util.CoalesceSlice(batch.Notes, batch.SermonNotes)

	return u.upsertSermonNotes(ctx, notes)
}

func (u *SyncUsecase) upsertGatherings(
	ctx context.Context,
	gatherings []*entity.Gathering,
) error {
	opt := lastWriteWinsOption[entity.Gathering]{
		id:          func(g *entity.Gathering) string { return g.ID },
		updatedAt:   func(g *entity.Gathering) time.Time { return g.UpdatedAt },
		findByID:    u.gatheringRepository.FindByID,
		notFoundErr: entityerror.ErrGatheringNotFound,
		create:      u.gatheringRepository.Create,
		update:      u.gatheringRepository.Update,
	}

	return upsertLastWriteWins(ctx, gatherings, opt)
}

func (u *SyncUsecase) upsertSermonSeries(
	ctx context.Context,
	series []*entity.SermonSeries,
) error {
	opt := lastWriteWinsOption[entity.SermonSeries]{
		id:          func(s *entity.SermonSeries) string { return s.ID },
		updatedAt:   func(s *entity.SermonSeries) time.Time { return s.UpdatedAt },
		findByID:    u.sermonSeriesRepository.FindByID,
		notFoundErr: entityerror.ErrSermonSeriesNotFound,
		create:      u.sermonSeriesRepository.Create,
		update:      u.sermonSeriesRepository.Update,
	}

	return upsertLastWriteWins(ctx, series, opt)
}

func (u *SyncUsecase) upsertPreachers(
	ctx context.Context,
	preachers []*entity.Preacher,
) error {
	opt := lastWriteWinsOption[entity.Preacher]{
		id:          func(p *entity.Preacher) string { return p.ID },
		updatedAt:   func(p *entity.Preacher) time.Time { return p.UpdatedAt },
		findByID:    u.preacherRepository.FindByID,
		notFoundErr: entityerror.ErrPreacherNotFound,
		create:      u.preacherRepository.Create,
		update:      u.preacherRepository.Update,
	}

	return upsertLastWriteWins(ctx, preachers, opt)
}

func (u *SyncUsecase) upsertTags(
	ctx context.Context,
	tags []*entity.Tag,
) error {
	opt := lastWriteWinsOption[entity.Tag]{
		id:          func(t *entity.Tag) string { return t.ID },
		updatedAt:   func(t *entity.Tag) time.Time { return t.UpdatedAt },
		findByID:    u.tagRepository.FindByID,
		notFoundErr: entityerror.ErrTagNotFound,
		create:      u.tagRepository.Create,
		update:      u.tagRepository.Update,
	}

	return upsertLastWriteWins(ctx, tags, opt)
}

func (u *SyncUsecase) upsertSermonNotes(
	ctx context.Context,
	notes []*entity.SermonNote,
) ([]string, []string, error) {
	missingContentIDs := make([]string, 0)
	conflicts := make([]string, 0)

	for _, note := range notes {
		missing, conflict, err := u.upsertSermonNote(ctx, note)
		if err != nil {
			return nil, nil, err
		}

		if conflict {
			conflicts = append(conflicts, note.ID)
			continue
		}

		if missing {
			missingContentIDs = append(missingContentIDs, note.ID)
		}
	}

	return missingContentIDs, conflicts, nil
}

func (u *SyncUsecase) upsertSermonNote(
	ctx context.Context,
	note *entity.SermonNote,
) (bool, bool, error) {
	if note == nil {
		return false, false, nil
	}

	var existing *entity.SermonNote
	if note.ID != "" {
		var err error
		existing, err = u.sermonNoteRepository.FindByID(ctx, note.ID)
		if err != nil &&
			!errors.Is(err, entityerror.ErrSermonNoteNotFound(note.ID)) {
			return false, false, err
		}
	}

	if existing == nil {
		if err := u.sermonNoteRepository.Create(ctx, note); err != nil {
			return false, false, err
		}

		return true, false, nil
	}

	if existing.HasConflictWith(note) {
		return false, true, nil
	}

	missing := existing.IsContentMissingFor(note)

	if err := u.sermonNoteRepository.Update(ctx, note); err != nil {
		return false, false, err
	}

	return missing, false, nil
}

type lastWriteWinsOption[T any] struct {
	id          func(*T) string
	updatedAt   func(*T) time.Time
	findByID    func(context.Context, string) (*T, error)
	notFoundErr func(string) *apperror.AppError
	create      func(context.Context, *T) error
	update      func(context.Context, *T) error
}

func upsertLastWriteWins[T any](
	ctx context.Context,
	items []*T,
	opt lastWriteWinsOption[T],
) error {
	for _, item := range items {
		if err := upsertLastWriteWinsItem(ctx, item, opt); err != nil {
			return err
		}
	}

	return nil
}

func upsertLastWriteWinsItem[T any](
	ctx context.Context,
	item *T,
	opt lastWriteWinsOption[T],
) error {
	if item == nil {
		return nil
	}

	id := opt.id(item)
	if id == "" {
		return opt.create(ctx, item)
	}

	existing, err := opt.findByID(ctx, id)
	if err != nil && !errors.Is(err, opt.notFoundErr(id)) {
		return err
	}

	if existing == nil {
		return opt.create(ctx, item)
	}

	if opt.updatedAt(item).Before(opt.updatedAt(existing)) {
		return nil
	}

	return opt.update(ctx, item)
}
