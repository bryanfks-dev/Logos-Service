package usecase

import (
	"context"
	"database/sql"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
)

const maxPushTxRetries = 6

type SyncPushChangesUsecase struct {
	txManager              *db.TxManager
	sermonNoteRepository   repository.SermonNoteRepository
	gatheringRepository    repository.GatheringRepository
	sermonSeriesRepository repository.SermonSeriesRepository
	preacherRepository     repository.PreacherRepository
	tagRepository          repository.TagRepository
	log                    *zap.Logger
}

type SyncPushChangesUsecaseOption struct {
	TxManager              *db.TxManager
	SermonNoteRepository   repository.SermonNoteRepository
	GatheringRepository    repository.GatheringRepository
	SermonSeriesRepository repository.SermonSeriesRepository
	PreacherRepository     repository.PreacherRepository
	TagRepository          repository.TagRepository
	Log                    *zap.Logger
}

func NewSyncPushChangesUsecase(
	opt *SyncPushChangesUsecaseOption,
) *SyncPushChangesUsecase {
	return &SyncPushChangesUsecase{
		txManager:              opt.TxManager,
		sermonNoteRepository:   opt.SermonNoteRepository,
		gatheringRepository:    opt.GatheringRepository,
		sermonSeriesRepository: opt.SermonSeriesRepository,
		preacherRepository:     opt.PreacherRepository,
		tagRepository:          opt.TagRepository,
		log:                    opt.Log,
	}
}

func (u *SyncPushChangesUsecase) Call(
	ctx context.Context,
	batch *model.SyncChangesBatch,
) (*model.SyncPushChangesResult, error) {
	if batch == nil {
		batch = &model.SyncChangesBatch{}
	}

	syncedAt := time.Now().UTC()

	var (
		missingContentIDs []string
		conflicts         []string
		err               error
	)

	txOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}

	for attempt := range maxPushTxRetries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		txFn := func(txCtx context.Context) error {
			var txErr error
			missingContentIDs, conflicts, txErr = u.applyPushBatch(
				txCtx,
				batch,
			)

			return txErr
		}

		err = u.txManager.DoWithOptions(ctx, txOpts, txFn)
		if err == nil || !isRetryableSyncError(err) ||
			attempt == maxPushTxRetries-1 {
			break
		}

		backoff := time.Duration(attempt+1)*10*time.Millisecond +
			time.Duration(rand.IntN(20))*time.Millisecond
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	if err != nil {
		u.log.Error("failed to push sync batch", zap.Error(err))
		return nil, err
	}

	result := &model.SyncPushChangesResult{
		SyncedAt:          syncedAt,
		MissingContentIDs: missingContentIDs,
		Conflicts:         conflicts,
	}

	return result, nil
}

func isRetryableSyncError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "40001") ||
		strings.Contains(msg, "40p01") ||
		strings.Contains(msg, "serialization") ||
		strings.Contains(msg, "deadlock") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database table is locked") ||
		strings.Contains(msg, "sqlite_busy") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "duplicate key")
}

func (u *SyncPushChangesUsecase) applyPushBatch(
	ctx context.Context,
	batch *model.SyncChangesBatch,
) ([]string, []string, error) {
	if err := u.upsertGatherings(ctx, batch.Gatherings); err != nil {
		return nil, nil, err
	}

	if err := u.upsertSermonSeries(ctx, batch.SermonSeries); err != nil {
		return nil, nil, err
	}

	if err := u.upsertPreachers(ctx, batch.Preachers); err != nil {
		return nil, nil, err
	}

	if err := u.upsertAndResolveTags(ctx, batch.Tags, batch.SermonNotes); err != nil {
		return nil, nil, err
	}

	return u.upsertSermonNotes(ctx, batch.SermonNotes)
}

func (u *SyncPushChangesUsecase) upsertGatherings(
	ctx context.Context,
	gatherings []*entity.Gathering,
) error {
	if len(gatherings) == 0 {
		return nil
	}

	ids := make([]string, 0, len(gatherings))
	seen := make(map[string]struct{}, len(gatherings))

	for _, gathering := range gatherings {
		if gathering == nil || gathering.ID == "" {
			continue
		}

		if _, ok := seen[gathering.ID]; ok {
			continue
		}

		seen[gathering.ID] = struct{}{}
		ids = append(ids, gathering.ID)
	}

	existingByID := make(map[string]*entity.Gathering, len(ids))

	if len(ids) > 0 {
		existingGatherings, err := u.gatheringRepository.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}

		for _, existing := range existingGatherings {
			if existing == nil {
				continue
			}

			existingByID[existing.ID] = existing
		}
	}

	for _, gathering := range gatherings {
		if err := u.upsertGathering(ctx, gathering, existingByID); err != nil {
			return err
		}
	}

	return nil
}

func (u *SyncPushChangesUsecase) upsertGathering(
	ctx context.Context,
	gathering *entity.Gathering,
	existingByID map[string]*entity.Gathering,
) error {
	if gathering == nil {
		return nil
	}

	if gathering.ID == "" {
		return u.gatheringRepository.Create(ctx, gathering)
	}

	existing := existingByID[gathering.ID]
	if existing == nil {
		if err := u.gatheringRepository.Create(ctx, gathering); err != nil {
			return err
		}

		existingByID[gathering.ID] = gathering

		return nil
	}

	if gathering.UpdatedAt.Before(existing.UpdatedAt) {
		return nil
	}

	if err := u.gatheringRepository.Update(ctx, gathering); err != nil {
		return err
	}

	existingByID[gathering.ID] = gathering

	return nil
}

func (u *SyncPushChangesUsecase) upsertSermonSeries(
	ctx context.Context,
	series []*entity.SermonSeries,
) error {
	if len(series) == 0 {
		return nil
	}

	ids := make([]string, 0, len(series))
	seen := make(map[string]struct{}, len(series))

	for _, item := range series {
		if item == nil || item.ID == "" {
			continue
		}

		if _, ok := seen[item.ID]; ok {
			continue
		}

		seen[item.ID] = struct{}{}
		ids = append(ids, item.ID)
	}

	existingByID := make(map[string]*entity.SermonSeries, len(ids))

	if len(ids) > 0 {
		existingSeries, err := u.sermonSeriesRepository.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}

		for _, existing := range existingSeries {
			if existing == nil {
				continue
			}

			existingByID[existing.ID] = existing
		}
	}

	for _, item := range series {
		if err := u.upsertSermonSeriesItem(ctx, item, existingByID); err != nil {
			return err
		}
	}

	return nil
}

func (u *SyncPushChangesUsecase) upsertSermonSeriesItem(
	ctx context.Context,
	series *entity.SermonSeries,
	existingByID map[string]*entity.SermonSeries,
) error {
	if series == nil {
		return nil
	}

	if series.ID == "" {
		return u.sermonSeriesRepository.Create(ctx, series)
	}

	existing := existingByID[series.ID]
	if existing == nil {
		if err := u.sermonSeriesRepository.Create(ctx, series); err != nil {
			return err
		}

		existingByID[series.ID] = series

		return nil
	}

	if series.UpdatedAt.Before(existing.UpdatedAt) {
		return nil
	}

	if err := u.sermonSeriesRepository.Update(ctx, series); err != nil {
		return err
	}

	existingByID[series.ID] = series

	return nil
}

func (u *SyncPushChangesUsecase) upsertPreachers(
	ctx context.Context,
	preachers []*entity.Preacher,
) error {
	if len(preachers) == 0 {
		return nil
	}

	ids := make([]string, 0, len(preachers))
	seen := make(map[string]struct{}, len(preachers))

	for _, preacher := range preachers {
		if preacher == nil || preacher.ID == "" {
			continue
		}

		if _, ok := seen[preacher.ID]; ok {
			continue
		}

		seen[preacher.ID] = struct{}{}
		ids = append(ids, preacher.ID)
	}

	existingByID := make(map[string]*entity.Preacher, len(ids))

	if len(ids) > 0 {
		existingPreachers, err := u.preacherRepository.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}

		for _, existing := range existingPreachers {
			if existing == nil {
				continue
			}

			existingByID[existing.ID] = existing
		}
	}

	for _, preacher := range preachers {
		if err := u.upsertPreacher(ctx, preacher, existingByID); err != nil {
			return err
		}
	}

	return nil
}

func (u *SyncPushChangesUsecase) upsertPreacher(
	ctx context.Context,
	preacher *entity.Preacher,
	existingByID map[string]*entity.Preacher,
) error {
	if preacher == nil {
		return nil
	}

	if preacher.ID == "" {
		return u.preacherRepository.Create(ctx, preacher)
	}

	existing := existingByID[preacher.ID]
	if existing == nil {
		if err := u.preacherRepository.Create(ctx, preacher); err != nil {
			return err
		}

		existingByID[preacher.ID] = preacher

		return nil
	}

	if preacher.UpdatedAt.Before(existing.UpdatedAt) {
		return nil
	}

	if err := u.preacherRepository.Update(ctx, preacher); err != nil {
		return err
	}

	existingByID[preacher.ID] = preacher

	return nil
}

type tagIndex struct {
	byID    map[string]*entity.Tag
	byName  map[string]*entity.Tag
	idAlias map[string]string
}

func newTagIndex(existing []*entity.Tag) *tagIndex {
	idx := &tagIndex{
		byID:    make(map[string]*entity.Tag, len(existing)),
		byName:  make(map[string]*entity.Tag, len(existing)),
		idAlias: make(map[string]string),
	}

	for _, tag := range existing {
		idx.register(tag)
	}

	return idx
}

func (idx *tagIndex) register(tag *entity.Tag) {
	if tag == nil {
		return
	}

	if tag.ID != "" {
		idx.byID[tag.ID] = tag
	}

	if tag.Name != "" {
		idx.byName[tag.Name] = tag
	}
}

func (idx *tagIndex) findMatch(id string, name string) *entity.Tag {
	if id != "" {
		if aliasedID, ok := idx.idAlias[id]; ok {
			if existing, found := idx.byID[aliasedID]; found {
				return existing
			}
		}

		if existing, found := idx.byID[id]; found {
			return existing
		}
	}

	if name != "" {
		if existing, found := idx.byName[name]; found {
			return existing
		}
	}

	return nil
}

func collectTagLookupKeys(
	tags []*entity.Tag,
	notes []*entity.SermonNote,
) ([]string, []string) {
	ids := make([]string, 0, len(tags))
	names := make([]string, 0, len(tags))
	seenIDs := make(map[string]struct{}, len(tags))
	seenNames := make(map[string]struct{}, len(tags))

	addTagKey := func(id string, name string) {
		if id != "" {
			if _, ok := seenIDs[id]; !ok {
				seenIDs[id] = struct{}{}
				ids = append(ids, id)
			}
		}

		if name != "" {
			if _, ok := seenNames[name]; !ok {
				seenNames[name] = struct{}{}
				names = append(names, name)
			}
		}
	}

	for _, tag := range tags {
		if tag == nil {
			continue
		}

		addTagKey(tag.ID, tag.Name)
	}

	for _, note := range notes {
		if note == nil {
			continue
		}

		for _, tag := range note.Tags {
			addTagKey(tag.ID, tag.Name)
		}
	}

	return ids, names
}

func (u *SyncPushChangesUsecase) upsertAndResolveTags(
	ctx context.Context,
	tags []*entity.Tag,
	notes []*entity.SermonNote,
) error {
	ids, names := collectTagLookupKeys(tags, notes)
	if len(ids) == 0 && len(names) == 0 {
		return nil
	}

	existingTags, err := u.tagRepository.FindByIDsOrNames(ctx, ids, names)
	if err != nil {
		return err
	}

	idx := newTagIndex(existingTags)

	for _, tag := range tags {
		if err := u.upsertTag(ctx, tag, idx); err != nil {
			return err
		}
	}

	for _, note := range notes {
		if err := u.resolveNoteTags(ctx, note, idx); err != nil {
			return err
		}
	}

	return nil
}

func (u *SyncPushChangesUsecase) upsertTag(
	ctx context.Context,
	tag *entity.Tag,
	idx *tagIndex,
) error {
	if tag == nil {
		return nil
	}

	existing := idx.findMatch(tag.ID, tag.Name)
	if existing == nil {
		if err := u.tagRepository.Create(ctx, tag); err != nil {
			return err
		}

		idx.register(tag)

		return nil
	}

	if tag.ID != "" && tag.ID != existing.ID {
		idx.idAlias[tag.ID] = existing.ID
	}

	if tag.UpdatedAt.Before(existing.UpdatedAt) {
		return nil
	}

	updated := *tag
	updated.ID = existing.ID

	// Avoid colliding with another tag that already owns updated.Name.
	if updated.Name != "" {
		if owner, ok := idx.byName[updated.Name]; ok &&
			owner.ID != existing.ID {
			updated.Name = existing.Name
		}
	}

	if err := u.tagRepository.Update(ctx, &updated); err != nil {
		return err
	}

	if existing.Name != "" && existing.Name != updated.Name {
		delete(idx.byName, existing.Name)
	}

	idx.register(&updated)

	return nil
}

func (u *SyncPushChangesUsecase) resolveNoteTags(
	ctx context.Context,
	note *entity.SermonNote,
	idx *tagIndex,
) error {
	if note == nil || len(note.Tags) == 0 {
		return nil
	}

	resolved := make([]entity.Tag, 0, len(note.Tags))
	seenIDs := make(map[string]struct{}, len(note.Tags))

	for _, noteTag := range note.Tags {
		canonical := idx.findMatch(noteTag.ID, noteTag.Name)
		if canonical == nil {
			if noteTag.Name == "" {
				continue
			}

			created := entity.Tag{
				ID:   noteTag.ID,
				Name: noteTag.Name,
			}
			if err := u.tagRepository.Create(ctx, &created); err != nil {
				return err
			}

			idx.register(&created)
			canonical = &created
		}

		if _, ok := seenIDs[canonical.ID]; ok {
			continue
		}

		seenIDs[canonical.ID] = struct{}{}
		resolved = append(resolved, *canonical)
	}

	note.Tags = resolved

	return nil
}

func (u *SyncPushChangesUsecase) upsertSermonNotes(
	ctx context.Context,
	notes []*entity.SermonNote,
) ([]string, []string, error) {
	missingContentIDs := make([]string, 0)
	conflicts := make([]string, 0)
	if len(notes) == 0 {
		return missingContentIDs, conflicts, nil
	}

	ids := make([]string, 0, len(notes))
	seen := make(map[string]struct{}, len(notes))

	for _, note := range notes {
		if note == nil || note.ID == "" {
			continue
		}

		if _, ok := seen[note.ID]; ok {
			continue
		}

		seen[note.ID] = struct{}{}
		ids = append(ids, note.ID)
	}

	existingByID := make(map[string]*entity.SermonNote, len(ids))

	if len(ids) > 0 {
		existingNotes, err := u.sermonNoteRepository.FindByIDsUnscoped(ctx, ids)
		if err != nil {
			return nil, nil, err
		}

		for _, existing := range existingNotes {
			if existing == nil {
				continue
			}

			existingByID[existing.ID] = existing
		}
	}

	for _, note := range notes {
		missing, conflict, err := u.upsertSermonNote(ctx, note, existingByID)
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

func (u *SyncPushChangesUsecase) upsertSermonNote(
	ctx context.Context,
	note *entity.SermonNote,
	existingByID map[string]*entity.SermonNote,
) (bool, bool, error) {
	if note == nil {
		return false, false, nil
	}

	var existing *entity.SermonNote
	if note.ID != "" {
		existing = existingByID[note.ID]
	}

	if existing == nil {
		if err := u.sermonNoteRepository.Create(ctx, note); err != nil {
			return false, false, err
		}

		if note.ID != "" {
			existingByID[note.ID] = note
		}

		return true, false, nil
	}

	if existing.HasConflictWith(note) {
		return false, true, nil
	}

	missing := existing.IsContentMissingFor(note)

	if note.FilePath == "" && existing.FilePath != "" {
		note.FilePath = existing.FilePath
	}

	if note.CreatedAt.IsZero() && !existing.CreatedAt.IsZero() {
		note.CreatedAt = existing.CreatedAt
	}

	if err := u.sermonNoteRepository.Update(ctx, note); err != nil {
		return false, false, err
	}

	existingByID[note.ID] = note

	return missing, false, nil
}
