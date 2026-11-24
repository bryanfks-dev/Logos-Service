package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/bryanfks-dev/selah/infrastructure/db"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/model/converter"
	"github.com/bryanfks-dev/selah/internal/repository"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SyncUsecase interface {
	SyncNotes(
		ctx context.Context,
		req *model.SyncNotesRequest,
	) (*model.SyncNotesResponse, error)
}

type syncUsecase struct {
	txManager      *db.TxManager
	sermonNoteRepo repository.SermonNoteRepository
	tagRepo        repository.TagRepository
	syncOutboxRepo repository.SyncOutboxRepository
	log            *zap.Logger
}

func NewSyncUsecase(
	txManager *db.TxManager,
	sermonNoteRepo repository.SermonNoteRepository,
	tagRepo repository.TagRepository,
	syncOutboxRepo repository.SyncOutboxRepository,
	log *zap.Logger,
) SyncUsecase {
	return &syncUsecase{
		txManager:      txManager,
		sermonNoteRepo: sermonNoteRepo,
		tagRepo:        tagRepo,
		syncOutboxRepo: syncOutboxRepo,
		log:            log,
	}
}

func (u *syncUsecase) SyncNotes(
	ctx context.Context,
	req *model.SyncNotesRequest,
) (*model.SyncNotesResponse, error) {
	if req == nil {
		req = &model.SyncNotesRequest{}
	}

	serverNow := time.Now().UTC()
	appliedIDs := make([]string, 0, len(req.Notes))

	err := u.txManager.Do(ctx, func(txCtx context.Context) error {
		for i := range req.Notes {
			item := &req.Notes[i]
			if item.ID == "" {
				item.ID = uuid.New().String()
			}

			existingNote, findErr := u.sermonNoteRepo.FindByIDUnscoped(
				txCtx,
				item.ID,
			)
			if findErr != nil &&
				!errors.Is(
					findErr,
					entityerror.ErrSermonNoteNotFound(item.ID),
				) {
				return findErr
			}

			// Case 1: Note does not exist on server
			if existingNote == nil {
				if item.IsDeleted {
					appliedIDs = append(appliedIDs, item.ID)
					continue
				}

				noteEntity := converter.SyncItemToNoteEntity(item)
				u.resolveTags(txCtx, &noteEntity, item.Tags)

				if err := u.sermonNoteRepo.Create(txCtx, &noteEntity); err != nil {
					return err
				}

				_ = u.syncOutboxRepo.Enqueue(txCtx, &entity.SyncOutboxEntry{
					EntityType: "sermon_notes",
					EntityID:   noteEntity.ID,
					Operation:  "create",
					QueuedAt:   serverNow,
				})

				appliedIDs = append(appliedIDs, item.ID)
				continue
			}

			// Case 2: Note exists on server - check conflict
			if item.Version < existingNote.Version {
				u.log.Warn("sync conflict detected; server version wins",
					zap.String("note_id", item.ID),
					zap.Int("client_version", item.Version),
					zap.Int("server_version", existingNote.Version),
				)
				continue
			}

			// Client version is >= server version; apply client change
			if item.IsDeleted {
				if !existingNote.DeletedAt.Valid {
					if err := u.sermonNoteRepo.Delete(txCtx, item.ID); err != nil {
						return err
					}

					_ = u.syncOutboxRepo.Enqueue(txCtx, &entity.SyncOutboxEntry{
						EntityType: "sermon_notes",
						EntityID:   item.ID,
						Operation:  "delete",
						QueuedAt:   serverNow,
					})
				}
				appliedIDs = append(appliedIDs, item.ID)
				continue
			}

			if existingNote.DeletedAt.Valid {
				if err := u.sermonNoteRepo.Restore(txCtx, item.ID); err != nil {
					return err
				}
			}

			noteEntity := converter.SyncItemToNoteEntity(item)
			noteEntity.Version = existingNote.Version + 1
			u.resolveTags(txCtx, &noteEntity, item.Tags)

			if err := u.sermonNoteRepo.Update(txCtx, &noteEntity); err != nil {
				return err
			}

			_ = u.syncOutboxRepo.Enqueue(txCtx, &entity.SyncOutboxEntry{
				EntityType: "sermon_notes",
				EntityID:   noteEntity.ID,
				Operation:  "update",
				QueuedAt:   serverNow,
			})

			appliedIDs = append(appliedIDs, item.ID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	filter := entity.SermonNoteFilter{
		IncludeDeleted: true,
		UpdatedSince:   req.LastSyncedAt,
	}

	serverNotes, err := u.sermonNoteRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	pulledItems := make([]model.SyncNoteItem, 0, len(serverNotes))
	for i := range serverNotes {
		pulledItems = append(
			pulledItems,
			converter.NoteEntityToSyncItem(&serverNotes[i]),
		)
	}

	return &model.SyncNotesResponse{
		SyncedAt:   serverNow,
		AppliedIDs: appliedIDs,
		Notes:      pulledItems,
	}, nil
}

func (u *syncUsecase) resolveTags(
	ctx context.Context,
	note *entity.SermonNote,
	tagNames []string,
) {
	if len(tagNames) == 0 {
		return
	}

	existingTags, err := u.tagRepo.FindAll(ctx)
	if err != nil {
		return
	}

	tagMap := make(map[string]entity.Tag, len(existingTags))
	for _, t := range existingTags {
		tagMap[t.Name] = t
	}

	resolvedTags := make([]entity.Tag, 0, len(tagNames))
	for _, name := range tagNames {
		if t, ok := tagMap[name]; ok {
			resolvedTags = append(resolvedTags, t)
			continue
		}

		newTag := entity.Tag{Name: name}
		if err := u.tagRepo.Create(ctx, &newTag); err == nil {
			tagMap[name] = newTag
			resolvedTags = append(resolvedTags, newTag)
		}
	}

	note.Tags = resolvedTags
}
