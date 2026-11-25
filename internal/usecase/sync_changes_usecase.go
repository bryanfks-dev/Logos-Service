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
	"github.com/bryanfks-dev/logos/pkg/util"
	"go.uber.org/zap"
)

const maxPushTxRetries = 6

type SyncChangesUsecase struct {
	txManager              *db.TxManager
	sermonNoteRepository   repository.SermonNoteRepository
	gatheringRepository    repository.GatheringRepository
	sermonSeriesRepository repository.SermonSeriesRepository
	preacherRepository     repository.PreacherRepository
	tagRepository          repository.TagRepository
	syncOutboxRepository   repository.SyncOutboxRepository
	log                    *zap.Logger
}

type SyncChangesUsecaseOption struct {
	TxManager              *db.TxManager
	SermonNoteRepository   repository.SermonNoteRepository
	GatheringRepository    repository.GatheringRepository
	SermonSeriesRepository repository.SermonSeriesRepository
	PreacherRepository     repository.PreacherRepository
	TagRepository          repository.TagRepository
	SyncOutboxRepository   repository.SyncOutboxRepository
	Log                    *zap.Logger
}

func NewSyncChangesUsecase(opt *SyncChangesUsecaseOption) *SyncChangesUsecase {
	return &SyncChangesUsecase{
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

func (u *SyncChangesUsecase) Pull(
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

	notes = util.NonNilSlice(notes)
	gatherings = util.NonNilSlice(gatherings)
	series = util.NonNilSlice(series)
	preachers = util.NonNilSlice(preachers)
	tags = util.NonNilSlice(tags)

	result := &model.SyncPullChangesResult{
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

func (u *SyncChangesUsecase) Push(
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

func (u *SyncChangesUsecase) applyPushBatch(
	ctx context.Context,
	batch *model.SyncChangesBatch,
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

	notes := util.CoalesceSlice(batch.Notes, batch.SermonNotes)

	if err := u.upsertAndResolveTags(ctx, batch.Tags, notes); err != nil {
		return nil, nil, err
	}

	return u.upsertSermonNotes(ctx, notes)
}

func (u *SyncChangesUsecase) upsertGatherings(
	ctx context.Context,
	gatherings []*entity.Gathering,
) error {
	opt := lastWriteWinsOption[entity.Gathering]{
		id:        func(g *entity.Gathering) string { return g.ID },
		updatedAt: func(g *entity.Gathering) time.Time { return g.UpdatedAt },
		findByIDs: u.gatheringRepository.FindByIDs,
		create:    u.gatheringRepository.Create,
		update:    u.gatheringRepository.Update,
	}

	return upsertLastWriteWins(ctx, gatherings, opt)
}

func (u *SyncChangesUsecase) upsertSermonSeries(
	ctx context.Context,
	series []*entity.SermonSeries,
) error {
	opt := lastWriteWinsOption[entity.SermonSeries]{
		id:        func(s *entity.SermonSeries) string { return s.ID },
		updatedAt: func(s *entity.SermonSeries) time.Time { return s.UpdatedAt },
		findByIDs: u.sermonSeriesRepository.FindByIDs,
		create:    u.sermonSeriesRepository.Create,
		update:    u.sermonSeriesRepository.Update,
	}

	return upsertLastWriteWins(ctx, series, opt)
}

func (u *SyncChangesUsecase) upsertPreachers(
	ctx context.Context,
	preachers []*entity.Preacher,
) error {
	opt := lastWriteWinsOption[entity.Preacher]{
		id:        func(p *entity.Preacher) string { return p.ID },
		updatedAt: func(p *entity.Preacher) time.Time { return p.UpdatedAt },
		findByIDs: u.preacherRepository.FindByIDs,
		create:    u.preacherRepository.Create,
		update:    u.preacherRepository.Update,
	}

	return upsertLastWriteWins(ctx, preachers, opt)
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

func (u *SyncChangesUsecase) upsertAndResolveTags(
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

func (u *SyncChangesUsecase) upsertTag(
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

func (u *SyncChangesUsecase) resolveNoteTags(
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

func (u *SyncChangesUsecase) upsertSermonNotes(
	ctx context.Context,
	notes []*entity.SermonNote,
) ([]string, []string, error) {
	missingContentIDs := make([]string, 0)
	conflicts := make([]string, 0)

	noteIDFn := func(n *entity.SermonNote) string { return n.ID }

	existingByID, err := fetchExistingByID(
		ctx,
		notes,
		noteIDFn,
		u.sermonNoteRepository.FindByIDsUnscoped,
	)
	if err != nil {
		return nil, nil, err
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

func (u *SyncChangesUsecase) upsertSermonNote(
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

type lastWriteWinsOption[T any] struct {
	id        func(*T) string
	updatedAt func(*T) time.Time
	findByIDs func(context.Context, []string) ([]*T, error)
	create    func(context.Context, *T) error
	update    func(context.Context, *T) error
}

func fetchExistingByID[T any](
	ctx context.Context,
	items []*T,
	getID func(*T) string,
	findByIDs func(context.Context, []string) ([]*T, error),
) (map[string]*T, error) {
	existingByID := make(map[string]*T)
	if len(items) == 0 {
		return existingByID, nil
	}

	ids := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		if item == nil {
			continue
		}

		id := getID(item)
		if id == "" {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return existingByID, nil
	}

	existingItems, err := findByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	for _, existing := range existingItems {
		if existing == nil {
			continue
		}

		existingByID[getID(existing)] = existing
	}

	return existingByID, nil
}

func upsertLastWriteWins[T any](
	ctx context.Context,
	items []*T,
	opt lastWriteWinsOption[T],
) error {
	existingByID, err := fetchExistingByID(ctx, items, opt.id, opt.findByIDs)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := upsertLastWriteWinsItem(ctx, item, existingByID, opt); err != nil {
			return err
		}
	}

	return nil
}

func upsertLastWriteWinsItem[T any](
	ctx context.Context,
	item *T,
	existingByID map[string]*T,
	opt lastWriteWinsOption[T],
) error {
	if item == nil {
		return nil
	}

	id := opt.id(item)
	if id == "" {
		return opt.create(ctx, item)
	}

	existing := existingByID[id]
	if existing == nil {
		if err := opt.create(ctx, item); err != nil {
			return err
		}

		existingByID[id] = item

		return nil
	}

	if opt.updatedAt(item).Before(opt.updatedAt(existing)) {
		return nil
	}

	if err := opt.update(ctx, item); err != nil {
		return err
	}

	existingByID[id] = item

	return nil
}
