package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"
	"uuid"

	"github.com/bryanfks-dev/selah/db"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/repository"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
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

func NewSyncUsecase(
	txManager *db.TxManager,
	sermonNoteRepository repository.SermonNoteRepository,
	gatheringRepository repository.GatheringRepository,
	sermonSeriesRepository repository.SermonSeriesRepository,
	preacherRepository repository.PreacherRepository,
	tagRepository repository.TagRepository,
	syncOutboxRepository repository.SyncOutboxRepository,
	log *zap.Logger,
) *SyncUsecase {
	return &SyncUsecase{
		txManager:              txManager,
		sermonNoteRepository:   sermonNoteRepository,
		gatheringRepository:    gatheringRepository,
		sermonSeriesRepository: sermonSeriesRepository,
		preacherRepository:     preacherRepository,
		tagRepository:          tagRepository,
		syncOutboxRepository:   syncOutboxRepository,
		log:                    log,
	}
}

func (u *SyncUsecase) Pull(
	ctx context.Context,
	since int64,
) (*model.SyncPullResult, error) {
	serverNow := time.Now().UTC()
	serverTime := serverNow.UnixMilli()

	notes, err := u.sermonNoteRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		return nil, err
	}

	gatherings, err := u.gatheringRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		return nil, err
	}

	series, err := u.sermonSeriesRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		return nil, err
	}

	preachers, err := u.preacherRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		return nil, err
	}

	tags, err := u.tagRepository.FindUpdatedSince(ctx, since)
	if err != nil {
		return nil, err
	}

	if notes == nil {
		notes = make([]*entity.SermonNote, 0)
	}
	if gatherings == nil {
		gatherings = make([]*entity.Gathering, 0)
	}
	if series == nil {
		series = make([]*entity.SermonSeries, 0)
	}
	if preachers == nil {
		preachers = make([]*entity.Preacher, 0)
	}
	if tags == nil {
		tags = make([]*entity.Tag, 0)
	}

	return &model.SyncPullResult{
		ServerTime:   serverTime,
		Notes:        notes,
		SermonNotes:  notes,
		Gatherings:   gatherings,
		Series:       series,
		SermonSeries: series,
		Preachers:    preachers,
		Tags:         tags,
	}, nil
}

func (u *SyncUsecase) Push(
	ctx context.Context,
	batch *model.SyncBatch,
) (*model.SyncPushResult, error) {
	if batch == nil {
		batch = &model.SyncBatch{}
	}

	serverNow := time.Now().UTC()
	syncedAt := serverNow.UnixMilli()
	missingContentIDs := make([]string, 0)
	conflicts := make([]string, 0)

	err := u.txManager.Do(ctx, func(txCtx context.Context) error {
		// 1. Upsert Gatherings using last-write-wins (UpdatedAt)
		for _, g := range batch.Gatherings {
			if g == nil {
				continue
			}

			if g.ID == "" {
				g.ID = uuid.NewV7().String()
			}

			existing, err := u.gatheringRepository.FindByID(txCtx, g.ID)
			if err != nil &&
				!errors.Is(err, entityerror.ErrGatheringNotFound(g.ID)) {
				return err
			}

			if existing == nil {
				if err := u.gatheringRepository.Create(txCtx, g); err != nil {
					return err
				}
			} else if !g.UpdatedAt.Before(existing.UpdatedAt) {
				if err := u.gatheringRepository.Update(txCtx, g); err != nil {
					return err
				}
			}
		}

		// 2. Upsert Series using last-write-wins (UpdatedAt)
		seriesList := batch.Series
		if len(seriesList) == 0 && len(batch.SermonSeries) > 0 {
			seriesList = batch.SermonSeries
		}

		for _, s := range seriesList {
			if s == nil {
				continue
			}

			if s.ID == "" {
				s.ID = uuid.NewV7().String()
			}

			existing, err := u.sermonSeriesRepository.FindByID(txCtx, s.ID)
			if err != nil &&
				!errors.Is(err, entityerror.ErrSermonSeriesNotFound(s.ID)) {
				return err
			}

			if existing == nil {
				if err := u.sermonSeriesRepository.Create(txCtx, s); err != nil {
					return err
				}
			} else if !s.UpdatedAt.Before(existing.UpdatedAt) {
				if err := u.sermonSeriesRepository.Update(txCtx, s); err != nil {
					return err
				}
			}
		}

		// 3. Upsert Preachers using last-write-wins (UpdatedAt)
		for _, p := range batch.Preachers {
			if p == nil {
				continue
			}

			if p.ID == "" {
				p.ID = uuid.NewV7().String()
			}

			existing, err := u.preacherRepository.FindByID(txCtx, p.ID)
			if err != nil &&
				!errors.Is(err, entityerror.ErrPreacherNotFound(p.ID)) {
				return err
			}

			if existing == nil {
				if err := u.preacherRepository.Create(txCtx, p); err != nil {
					return err
				}
			} else if !p.UpdatedAt.Before(existing.UpdatedAt) {
				if err := u.preacherRepository.Update(txCtx, p); err != nil {
					return err
				}
			}
		}

		// 4. Upsert Tags using last-write-wins (UpdatedAt)
		for _, t := range batch.Tags {
			if t == nil {
				continue
			}

			if t.ID == "" {
				t.ID = uuid.NewV7().String()
			}

			existing, err := u.tagRepository.FindByID(txCtx, t.ID)
			if err != nil && !errors.Is(err, entityerror.ErrTagNotFound(t.ID)) {
				return err
			}

			if existing == nil {
				if err := u.tagRepository.Create(txCtx, t); err != nil {
					return err
				}
			} else if !t.UpdatedAt.Before(existing.UpdatedAt) {
				if err := u.tagRepository.Update(txCtx, t); err != nil {
					return err
				}
			}
		}

		// 5. Upsert incoming SermonNotes
		notes := batch.Notes
		if len(notes) == 0 && len(batch.SermonNotes) > 0 {
			notes = batch.SermonNotes
		}

		for _, note := range notes {
			if note == nil {
				continue
			}

			if note.ID == "" {
				note.ID = uuid.NewV7().String()
			}

			existingNote, findErr := u.sermonNoteRepository.FindByID(
				txCtx,
				note.ID,
			)
			if findErr != nil &&
				!errors.Is(
					findErr,
					entityerror.ErrSermonNoteNotFound(note.ID),
				) {
				return findErr
			}

			if existingNote != nil && existingNote.Version > note.Version &&
				existingNote.ContentHash != note.ContentHash {
				conflicts = append(conflicts, note.ID)
				continue
			}

			isMissing := false
			if existingNote == nil {
				isMissing = true
			} else {
				storedHash := existingNote.ContentHash
				if existingNote.FilePath != "" {
					if data, err := os.ReadFile(existingNote.FilePath); err == nil {
						h := sha256.Sum256(data)
						storedHash = hex.EncodeToString(h[:])
					}
				}

				if note.ContentHash != storedHash {
					isMissing = true
				}
			}

			if isMissing {
				missingContentIDs = append(missingContentIDs, note.ID)
			}

			if existingNote == nil {
				if err := u.sermonNoteRepository.Create(txCtx, note); err != nil {
					return err
				}
			} else {
				if err := u.sermonNoteRepository.Update(txCtx, note); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &model.SyncPushResult{
		SyncedAt:          syncedAt,
		MissingContentIDs: missingContentIDs,
		Conflicts:         conflicts,
	}, nil
}
