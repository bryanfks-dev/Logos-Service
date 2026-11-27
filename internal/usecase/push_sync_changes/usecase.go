package pushsyncchanges

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	"go.uber.org/zap"
)

type PushSyncChangesInput struct {
	SermonNotes  []*PushSermonNoteInput
	Gatherings   []*PushGatheringInput
	SermonSeries []*PushSermonSeriesInput
	Preachers    []*PushPreacherInput
	Tags         []*PushTagInput
}

type PushSermonNoteInput struct {
	ID          string
	GatheringID *string
	SeriesID    *string
	PreacherID  *string
	Title       string
	SermonDate  time.Time
	FilePath    string
	ContentHash string
	PlainText   string
	ContentJSON string
	Version     int
	Tags        []PushNoteTagInput
	Verses      []PushNoteVerseInput
	IsDeleted   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PushNoteTagInput struct {
	ID   string
	Name string
}

type PushNoteVerseInput struct {
	ID          string
	Translation string
	BookID      int
	Chapter     int
	VerseStart  int
	VerseEnd    *int
	DisplayRef  string
}

type PushGatheringInput struct {
	ID        string
	Name      string
	ColorHex  string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PushSermonSeriesInput struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PushPreacherInput struct {
	ID         string
	Name       string
	ChurchName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PushTagInput struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PushSyncChangesResult struct {
	SyncedAt          time.Time
	MissingContentIDs []string
	Conflicts         []string
}

const maxPushTxRetries = 6

var retryableSyncErrorMarkers = []string{
	"40001",
	"40p01",
	"55p03",
	"23505",
	"could not serialize access",
	"serialization failure",
	"deadlock detected",
	"could not obtain lock",
	"lock timeout",
	"duplicate key value violates unique constraint",
}

func (u *PushSyncChangesUsecase) Call(
	ctx context.Context,
	input *PushSyncChangesInput,
) (*PushSyncChangesResult, error) {
	u.log.Info("push sync changes started")

	if input == nil {
		u.log.Debug("push sync changes received nil input")
		input = &PushSyncChangesInput{}
	}

	serverTime := time.Now().UTC()

	var (
		missingContentIDs []string
		conflicts         []string
		err               error
	)

	txOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}

	txFn := func(txCtx context.Context) error {
		var txErr error
		missingContentIDs, conflicts, txErr = u.applyPushBatch(txCtx, input)

		return txErr
	}

	for attempt := range maxPushTxRetries {
		if ctxErr := ctx.Err(); ctxErr != nil {
			u.log.Error("push sync changes context error", zap.Error(ctxErr))

			return nil, ctxErr
		}

		err = u.txManager.DoWithOptions(ctx, txOpts, txFn)
		if err == nil || !u.isRetryableSyncError(err) ||
			attempt == maxPushTxRetries-1 {
			break
		}

		u.log.Error(
			"retryable push sync transaction error",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
		)

		if waitErr := u.waitRetryBackoff(ctx, attempt); waitErr != nil {
			u.log.Error(
				"push sync changes retry backoff aborted",
				zap.Error(waitErr),
			)

			return nil, waitErr
		}
	}

	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.IsDomain() {
			return nil, err
		}

		u.log.Error("failed to push sync batch", zap.Error(err))

		return nil, err
	}

	result := &PushSyncChangesResult{
		SyncedAt:          serverTime,
		MissingContentIDs: missingContentIDs,
		Conflicts:         conflicts,
	}

	u.log.Info(
		"push sync changes completed",
		zap.Int("missing_content_count", len(missingContentIDs)),
		zap.Int("conflicts_count", len(conflicts)),
	)

	return result, nil
}

func (u *PushSyncChangesUsecase) applyPushBatch(
	ctx context.Context,
	input *PushSyncChangesInput,
) ([]string, []string, error) {
	gatherings := u.toGatheringEntities(input.Gatherings)
	if err := u.upsertGatherings(ctx, gatherings); err != nil {
		return nil, nil, err
	}

	series := u.toSermonSeriesEntities(input.SermonSeries)
	if err := u.upsertSermonSeries(ctx, series); err != nil {
		return nil, nil, err
	}

	preachers := u.toPreacherEntities(input.Preachers)
	if err := u.upsertPreachers(ctx, preachers); err != nil {
		return nil, nil, err
	}

	tags := u.toTagEntities(input.Tags)
	notes := u.toSermonNoteEntities(input.SermonNotes)

	if err := u.upsertAndResolveTags(ctx, tags, notes); err != nil {
		return nil, nil, err
	}

	return u.upsertSermonNotes(ctx, notes)
}

func (u *PushSyncChangesUsecase) toGatheringEntities(
	items []*PushGatheringInput,
) []*entity.Gathering {
	res := make([]*entity.Gathering, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		g := &entity.Gathering{
			ID:        item.ID,
			Name:      item.Name,
			ColorHex:  item.ColorHex,
			SortOrder: item.SortOrder,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		res = append(res, g)
	}

	return res
}

func (u *PushSyncChangesUsecase) toSermonSeriesEntities(
	items []*PushSermonSeriesInput,
) []*entity.SermonSeries {
	res := make([]*entity.SermonSeries, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		s := &entity.SermonSeries{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
		res = append(res, s)
	}

	return res
}

func (u *PushSyncChangesUsecase) toPreacherEntities(
	items []*PushPreacherInput,
) []*entity.Preacher {
	res := make([]*entity.Preacher, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		p := &entity.Preacher{
			ID:         item.ID,
			Name:       item.Name,
			ChurchName: item.ChurchName,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		}
		res = append(res, p)
	}

	return res
}

func (u *PushSyncChangesUsecase) toTagEntities(
	items []*PushTagInput,
) []*entity.Tag {
	res := make([]*entity.Tag, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		t := &entity.Tag{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		res = append(res, t)
	}

	return res
}

func (u *PushSyncChangesUsecase) toSermonNoteEntities(
	items []*PushSermonNoteInput,
) []*entity.SermonNote {
	res := make([]*entity.SermonNote, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		version := item.Version
		if version < 1 {
			version = 1
		}

		sermonDate := item.SermonDate
		if sermonDate.IsZero() {
			sermonDate = time.Now().UTC()
		}

		verses := make([]entity.NoteVerse, 0, len(item.Verses))
		for _, v := range item.Verses {
			verse := entity.NoteVerse{
				ID:          v.ID,
				NoteID:      item.ID,
				Translation: v.Translation,
				BookID:      v.BookID,
				Chapter:     v.Chapter,
				VerseStart:  v.VerseStart,
				VerseEnd:    v.VerseEnd,
				DisplayRef:  v.DisplayRef,
			}
			verses = append(verses, verse)
		}

		tags := make([]entity.Tag, 0, len(item.Tags))
		for _, t := range item.Tags {
			tag := entity.Tag{
				ID:   t.ID,
				Name: t.Name,
			}
			tags = append(tags, tag)
		}

		n := &entity.SermonNote{
			ID:          item.ID,
			GatheringID: item.GatheringID,
			SeriesID:    item.SeriesID,
			PreacherID:  item.PreacherID,
			Title:       item.Title,
			SermonDate:  sermonDate,
			FilePath:    item.FilePath,
			ContentHash: item.ContentHash,
			PlainText:   item.PlainText,
			ContentJSON: item.ContentJSON,
			Version:     version,
			Verses:      verses,
			Tags:        tags,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
		res = append(res, n)
	}

	return res
}

func (u *PushSyncChangesUsecase) isRetryableSyncError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	for _, marker := range retryableSyncErrorMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}

	return false
}

func (u *PushSyncChangesUsecase) waitRetryBackoff(
	ctx context.Context,
	attempt int,
) error {
	backoff := time.Duration(attempt+1)*10*time.Millisecond +
		time.Duration(rand.IntN(20))*time.Millisecond

	timer := time.NewTimer(backoff)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (u *PushSyncChangesUsecase) upsertGatherings(
	ctx context.Context,
	gatherings []*entity.Gathering,
) error {
	if len(gatherings) == 0 {
		u.log.Debug("no gatherings to upsert")

		return nil
	}

	upserter := &lastWriteWinsUpserter[entity.Gathering]{
		id:        func(g *entity.Gathering) string { return g.ID },
		updatedAt: func(g *entity.Gathering) time.Time { return g.UpdatedAt },
		beforeUpdate: func(incoming *entity.Gathering, existing *entity.Gathering) {
			if incoming.CreatedAt.IsZero() && !existing.CreatedAt.IsZero() {
				incoming.CreatedAt = existing.CreatedAt
			}
		},
		findByIDs:   u.gatheringRepository.FindByIDs,
		createBatch: u.gatheringRepository.CreateBatch,
		updateBatch: u.gatheringRepository.UpdateBatch,
		log:         u.log,
	}

	return upserter.upsert(ctx, gatherings)
}

func (u *PushSyncChangesUsecase) upsertSermonSeries(
	ctx context.Context,
	series []*entity.SermonSeries,
) error {
	if len(series) == 0 {
		u.log.Debug("no sermon series to upsert")

		return nil
	}

	upserter := &lastWriteWinsUpserter[entity.SermonSeries]{
		id:        func(s *entity.SermonSeries) string { return s.ID },
		updatedAt: func(s *entity.SermonSeries) time.Time { return s.UpdatedAt },
		beforeUpdate: func(incoming *entity.SermonSeries, existing *entity.SermonSeries) {
			if incoming.CreatedAt.IsZero() && !existing.CreatedAt.IsZero() {
				incoming.CreatedAt = existing.CreatedAt
			}
		},
		findByIDs:   u.sermonSeriesRepository.FindByIDs,
		createBatch: u.sermonSeriesRepository.CreateBatch,
		updateBatch: u.sermonSeriesRepository.UpdateBatch,
		log:         u.log,
	}

	return upserter.upsert(ctx, series)
}

func (u *PushSyncChangesUsecase) upsertPreachers(
	ctx context.Context,
	preachers []*entity.Preacher,
) error {
	if len(preachers) == 0 {
		u.log.Debug("no preachers to upsert")

		return nil
	}

	upserter := &lastWriteWinsUpserter[entity.Preacher]{
		id:        func(p *entity.Preacher) string { return p.ID },
		updatedAt: func(p *entity.Preacher) time.Time { return p.UpdatedAt },
		beforeUpdate: func(incoming *entity.Preacher, existing *entity.Preacher) {
			if incoming.CreatedAt.IsZero() && !existing.CreatedAt.IsZero() {
				incoming.CreatedAt = existing.CreatedAt
			}
		},
		findByIDs:   u.preacherRepository.FindByIDs,
		createBatch: u.preacherRepository.CreateBatch,
		updateBatch: u.preacherRepository.UpdateBatch,
		log:         u.log,
	}

	return upserter.upsert(ctx, preachers)
}

func (u *PushSyncChangesUsecase) upsertAndResolveTags(
	ctx context.Context,
	tags []*entity.Tag,
	notes []*entity.SermonNote,
) error {
	ids, names := u.collectTagLookupKeys(tags, notes)
	if len(ids) == 0 && len(names) == 0 {
		u.log.Debug("no tags to upsert or resolve")

		return nil
	}

	existingTags, err := u.tagRepository.FindByIDsOrNames(ctx, ids, names)
	if err != nil {
		return err
	}

	idx := u.newTagIndex(existingTags)
	planner := u.newTagBatchPlanner(idx)

	for _, tag := range tags {
		planner.planTag(tag)
	}

	for _, note := range notes {
		planner.planNoteTags(note)
	}

	if err := u.tagRepository.CreateBatch(ctx, planner.toCreate); err != nil {
		return err
	}

	for _, created := range planner.toCreate {
		idx.register(created)
	}

	if err := u.tagRepository.UpdateBatch(ctx, planner.toUpdate); err != nil {
		return err
	}

	for _, note := range notes {
		u.resolveNoteTags(note, idx)
	}

	return nil
}

func (u *PushSyncChangesUsecase) collectTagLookupKeys(
	tags []*entity.Tag,
	notes []*entity.SermonNote,
) ([]string, []string) {
	keys := u.newTagKeySet(len(tags))

	for _, tag := range tags {
		if tag == nil {
			continue
		}

		keys.add(tag.ID, tag.Name)
	}

	for _, note := range notes {
		if note == nil {
			continue
		}

		for _, tag := range note.Tags {
			keys.add(tag.ID, tag.Name)
		}
	}

	return keys.ids, keys.names
}

func (u *PushSyncChangesUsecase) resolveNoteTags(
	note *entity.SermonNote,
	idx *tagIndex,
) {
	if note == nil || len(note.Tags) == 0 {
		return
	}

	resolved := make([]entity.Tag, 0, len(note.Tags))
	seenIDs := make(map[string]struct{}, len(note.Tags))

	for _, noteTag := range note.Tags {
		canonical := idx.findMatch(noteTag.ID, noteTag.Name)
		if canonical == nil {
			continue
		}

		if _, ok := seenIDs[canonical.ID]; ok {
			continue
		}

		seenIDs[canonical.ID] = struct{}{}
		resolved = append(resolved, *canonical)
	}

	note.Tags = resolved
}

func (u *PushSyncChangesUsecase) upsertSermonNotes(
	ctx context.Context,
	notes []*entity.SermonNote,
) ([]string, []string, error) {
	missingContentIDs := make([]string, 0)
	conflicts := make([]string, 0)
	if len(notes) == 0 {
		u.log.Debug("no sermon notes to upsert")

		return missingContentIDs, conflicts, nil
	}

	ids := u.collectUniqueSermonNoteIDs(notes)

	existingByID, err := u.fetchExistingSermonNotes(ctx, ids)
	if err != nil {
		return nil, nil, err
	}

	toCreate := make([]*entity.SermonNote, 0)
	toUpdate := make([]*entity.SermonNote, 0)
	missingContentNotes := make([]*entity.SermonNote, 0, len(notes))
	pendingCreateByID := make(map[string]int, len(notes))
	pendingUpdateByID := make(map[string]int, len(notes))
	seenConflicts := make(map[string]struct{}, len(notes))

	for _, note := range notes {
		if note == nil {
			continue
		}

		var existing *entity.SermonNote
		if note.ID != "" {
			existing = existingByID[note.ID]
		}

		if existing == nil {
			if note.ID != "" {
				pendingCreateByID[note.ID] = len(toCreate)
				existingByID[note.ID] = note
			}

			toCreate = append(toCreate, note)
			missingContentNotes = append(missingContentNotes, note)

			continue
		}

		if existing.HasConflictWith(note) {
			u.log.Debug(
				"sermon note conflict detected",
				zap.String("note_id", note.ID),
				zap.Int("existing_version", existing.Version),
				zap.Int("incoming_version", note.Version),
			)
			conflicts = u.appendUniqueID(conflicts, seenConflicts, note.ID)

			continue
		}

		if existing.IsContentMissingFor(note) {
			u.log.Debug(
				"sermon note content missing for updated note",
				zap.String("note_id", note.ID),
			)
			missingContentNotes = append(missingContentNotes, note)
		}

		if note.FilePath == "" && existing.FilePath != "" {
			note.FilePath = existing.FilePath
		}

		if note.CreatedAt.IsZero() && !existing.CreatedAt.IsZero() {
			note.CreatedAt = existing.CreatedAt
		}

		if createIdx, inCreate := pendingCreateByID[note.ID]; inCreate {
			toCreate[createIdx] = note
		} else if updateIdx, inUpdate := pendingUpdateByID[note.ID]; inUpdate {
			toUpdate[updateIdx] = note
		} else {
			pendingUpdateByID[note.ID] = len(toUpdate)
			toUpdate = append(toUpdate, note)
		}

		existingByID[note.ID] = note
	}

	if err := u.sermonNoteRepository.CreateBatch(ctx, toCreate); err != nil {
		return nil, nil, err
	}

	if err := u.sermonNoteRepository.UpdateBatch(ctx, toUpdate); err != nil {
		return nil, nil, err
	}

	seenMissing := make(map[string]struct{}, len(missingContentNotes))
	for _, note := range missingContentNotes {
		if note == nil {
			continue
		}

		if _, isExisting := existingByID[note.ID]; !isExisting {
			u.log.Debug(
				"sermon note content missing for new note",
				zap.String("note_id", note.ID),
			)
		}

		missingContentIDs = u.appendUniqueID(
			missingContentIDs,
			seenMissing,
			note.ID,
		)
	}

	return missingContentIDs, conflicts, nil
}

func (u *PushSyncChangesUsecase) fetchExistingSermonNotes(
	ctx context.Context,
	ids []string,
) (map[string]*entity.SermonNote, error) {
	existingByID := make(map[string]*entity.SermonNote, len(ids))
	if len(ids) == 0 {
		return existingByID, nil
	}

	existingNotes, err := u.sermonNoteRepository.FindByIDsUnscoped(ctx, ids)
	if err != nil {
		return nil, err
	}

	for _, existing := range existingNotes {
		if existing == nil {
			continue
		}

		existingByID[existing.ID] = existing
	}

	return existingByID, nil
}

func (u *PushSyncChangesUsecase) collectUniqueSermonNoteIDs(
	notes []*entity.SermonNote,
) []string {
	ids := make([]string, 0, len(notes))
	seen := make(map[string]struct{}, len(notes))

	for _, note := range notes {
		if note == nil {
			continue
		}

		ids = u.appendUniqueID(ids, seen, note.ID)
	}

	return ids
}

func (u *PushSyncChangesUsecase) appendUniqueID(
	ids []string,
	seen map[string]struct{},
	id string,
) []string {
	if id == "" {
		return ids
	}

	if _, ok := seen[id]; ok {
		return ids
	}

	seen[id] = struct{}{}

	return append(ids, id)
}

func (u *PushSyncChangesUsecase) newTagIndex(existing []*entity.Tag) *tagIndex {
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

func (u *PushSyncChangesUsecase) newTagKeySet(capacity int) *tagKeySet {
	return &tagKeySet{
		ids:       make([]string, 0, capacity),
		names:     make([]string, 0, capacity),
		seenIDs:   make(map[string]struct{}, capacity),
		seenNames: make(map[string]struct{}, capacity),
	}
}

func (u *PushSyncChangesUsecase) newTagBatchPlanner(
	idx *tagIndex,
) *tagBatchPlanner {
	return &tagBatchPlanner{
		idx:               idx,
		toCreate:          make([]*entity.Tag, 0),
		toUpdate:          make([]*entity.Tag, 0),
		pendingCreateIdx:  make(map[*entity.Tag]int),
		pendingUpdateByID: make(map[string]int),
		log:               u.log,
	}
}
