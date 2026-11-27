package sermonnote

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *sermonNoteRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *sermonNoteRepository) ensureNoteFilePath(note *entity.SermonNote) {
	if note.ContentJSON == "" || note.FilePath != "" {
		return
	}

	note.FilePath = filepath.Join(
		os.TempDir(),
		"logos_notes",
		note.ID+".json",
	)
}

func (r *sermonNoteRepository) commitNoteFile(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	if note.ContentJSON == "" {
		return nil
	}

	r.ensureNoteFilePath(note)

	filePath := note.FilePath
	contentJSON := note.ContentJSON

	writeHook := func() error {
		return r.fsManager.WriteFileAtomic(filePath, []byte(contentJSON))
	}

	if db.AddAfterCommit(ctx, writeHook) {
		return nil
	}

	return r.fsManager.WriteFileAtomic(filePath, []byte(contentJSON))
}

func (r *sermonNoteRepository) hydrateNoteFile(note *entity.SermonNote) {
	if note.FilePath == "" {
		return
	}

	if data, err := os.ReadFile(note.FilePath); err == nil {
		note.ContentJSON = string(data)
	}
}

func (r *sermonNoteRepository) hydrateNoteRelations(
	ctx context.Context,
	note *entity.SermonNote,
) {
	var verses []entity.NoteVerse
	_ = r.conn(ctx).Where("note_id = ?", note.ID).Find(&verses).Error
	note.Verses = verses

	var tags []entity.Tag
	_ = r.conn(ctx).Table("tags").
		Joins("JOIN note_tags ON note_tags.tag_id = tags.id").
		Where("note_tags.note_id = ?", note.ID).
		Find(&tags).
		Error
	note.Tags = tags

	var fts entity.NotesFTS
	if err := r.conn(ctx).Where("note_id = ?", note.ID).First(&fts).Error; err == nil {
		note.FTS = &fts
	}

	r.hydrateNoteFile(note)
}

func (r *sermonNoteRepository) saveChildrenBatch(
	ctx context.Context,
	notes []*entity.SermonNote,
	deleteExisting bool,
) error {
	if len(notes) == 0 {
		return nil
	}

	db := r.conn(ctx)

	noteIDs := make([]string, 0, len(notes))
	allVerses := make([]entity.NoteVerse, 0)
	allNoteTags := make([]entity.NoteTag, 0)
	allFTS := make([]entity.NotesFTS, 0, len(notes))

	for _, note := range notes {
		if note == nil {
			continue
		}

		noteIDs = append(noteIDs, note.ID)

		for i := range note.Verses {
			note.Verses[i].NoteID = note.ID
			allVerses = append(allVerses, note.Verses[i])
		}

		seenTagIDs := make(map[string]struct{}, len(note.Tags))
		for _, tag := range note.Tags {
			if tag.ID == "" {
				continue
			}

			if _, ok := seenTagIDs[tag.ID]; ok {
				continue
			}

			seenTagIDs[tag.ID] = struct{}{}
			noteTag := entity.NoteTag{
				NoteID: note.ID,
				TagID:  tag.ID,
			}
			allNoteTags = append(allNoteTags, noteTag)
		}

		fts := note.FTS
		if fts == nil {
			fts = &entity.NotesFTS{
				Title:     note.Title,
				PlainText: note.PlainText,
			}
		}

		fts.NoteID = note.ID
		allFTS = append(allFTS, *fts)
	}

	if len(noteIDs) == 0 {
		return nil
	}

	if deleteExisting {
		noteVerseModel := &entity.NoteVerse{}
		if err := db.Where("note_id IN ?", noteIDs).Delete(noteVerseModel).Error; err != nil {
			return err
		}

		noteTagModel := &entity.NoteTag{}
		if err := db.Where("note_id IN ?", noteIDs).Delete(noteTagModel).Error; err != nil {
			return err
		}

		notesFTSModel := &entity.NotesFTS{}
		if err := db.Where("note_id IN ?", noteIDs).Delete(notesFTSModel).Error; err != nil {
			return err
		}
	}

	if len(allVerses) > 0 {
		if err := db.CreateInBatches(&allVerses, 500).Error; err != nil {
			return err
		}
	}

	if len(allNoteTags) > 0 {
		if err := db.CreateInBatches(&allNoteTags, 500).Error; err != nil {
			return err
		}
	}

	if len(allFTS) > 0 {
		if err := db.CreateInBatches(&allFTS, 500).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *sermonNoteRepository) Create(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	r.log.Info("create started", zap.Any("param", note))

	db := r.conn(ctx)

	if err := db.Omit(clause.Associations).Create(note).Error; err != nil {
		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	notes := []*entity.SermonNote{note}
	if err := r.saveChildrenBatch(ctx, notes, true); err != nil {
		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	if err := r.commitNoteFile(ctx, note); err != nil {
		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", note))

	return nil
}

func (r *sermonNoteRepository) CreateBatch(
	ctx context.Context,
	notes []*entity.SermonNote,
) error {
	r.log.Info("create batch started", zap.Int("count", len(notes)))

	if len(notes) == 0 {
		r.log.Debug("create batch skipped due to empty slice")

		return nil
	}

	db := r.conn(ctx)

	if err := db.Omit(clause.Associations).CreateInBatches(notes, 500).Error; err != nil {
		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	if err := r.saveChildrenBatch(ctx, notes, false); err != nil {
		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	for _, note := range notes {
		if note == nil {
			continue
		}

		if err := r.commitNoteFile(ctx, note); err != nil {
			return entityerror.ErrSermonNoteCreationFailed(err)
		}
	}

	r.log.Info("create batch completed", zap.Int("count", len(notes)))

	return nil
}

func (r *sermonNoteRepository) Update(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	r.log.Info("update started", zap.Any("param", note))

	r.ensureNoteFilePath(note)

	db := r.conn(ctx)

	if err := db.Omit(clause.Associations).Save(note).Error; err != nil {
		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	notes := []*entity.SermonNote{note}
	if err := r.saveChildrenBatch(ctx, notes, true); err != nil {
		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	if err := r.commitNoteFile(ctx, note); err != nil {
		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", note))

	return nil
}

func (r *sermonNoteRepository) UpdateBatch(
	ctx context.Context,
	notes []*entity.SermonNote,
) error {
	r.log.Info("update batch started", zap.Int("count", len(notes)))

	if len(notes) == 0 {
		r.log.Debug("update batch skipped due to empty slice")

		return nil
	}

	for _, note := range notes {
		if note == nil {
			continue
		}

		r.ensureNoteFilePath(note)
	}

	db := r.conn(ctx)
	onConflict := clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}

	if err := db.Omit(clause.Associations).Clauses(onConflict).CreateInBatches(notes, 500).Error; err != nil {
		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	if err := r.saveChildrenBatch(ctx, notes, true); err != nil {
		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	for _, note := range notes {
		if note == nil {
			continue
		}

		if err := r.commitNoteFile(ctx, note); err != nil {
			return entityerror.ErrSermonNoteUpdateFailed(err)
		}
	}

	r.log.Info("update batch completed", zap.Int("count", len(notes)))

	return nil
}

func (r *sermonNoteRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.SermonNote, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var note entity.SermonNote

	if err := r.conn(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonNoteNotFound(id)
		}

		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	r.hydrateNoteRelations(ctx, &note)

	r.log.Info("find by id completed", zap.Any("param", id))

	return &note, nil
}

func (r *sermonNoteRepository) FindByIDUnscoped(
	ctx context.Context,
	id string,
) (*entity.SermonNote, error) {
	r.log.Info("find by id unscoped started", zap.Any("param", id))

	var note entity.SermonNote

	if err := r.conn(ctx).Unscoped().Where("id = ?", id).First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonNoteNotFound(id)
		}

		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	r.hydrateNoteRelations(ctx, &note)

	r.log.Info("find by id unscoped completed", zap.Any("param", id))

	return &note, nil
}

func (r *sermonNoteRepository) FindByIDsUnscoped(
	ctx context.Context,
	ids []string,
) ([]*entity.SermonNote, error) {
	r.log.Info("find by ids unscoped started", zap.Strings("ids", ids))

	if len(ids) == 0 {
		r.log.Debug("find by ids unscoped skipped due to empty ids")

		return make([]*entity.SermonNote, 0), nil
	}

	var notes []*entity.SermonNote

	query := db.WithRowLock(
		ctx,
		r.conn(ctx).Unscoped().Where("id IN ?", ids),
	)
	if err := query.Find(&notes).Error; err != nil {
		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	for _, note := range notes {
		r.hydrateNoteFile(note)
	}

	r.log.Info("find by ids unscoped completed", zap.Int("count", len(notes)))

	return notes, nil
}

func (r *sermonNoteRepository) FindAll(
	ctx context.Context,
	filter entity.SermonNoteFilter,
) ([]*entity.SermonNote, error) {
	r.log.Info("find all started", zap.Any("param", filter))

	db := r.conn(ctx)
	if filter.IncludeDeleted {
		db = db.Unscoped()
	}

	sermonNoteModel := &entity.SermonNote{}
	query := db.Model(sermonNoteModel)

	if !filter.IncludeDeleted {
		query = query.Where("sermon_notes.deleted_at IS NULL")
	}

	if filter.UpdatedSince != nil {
		query = query.Where("sermon_notes.updated_at > ?", *filter.UpdatedSince)
	}

	if filter.GatheringID != nil {
		query = query.Where(
			"sermon_notes.gathering_id = ?",
			*filter.GatheringID,
		)
	}

	if filter.SeriesID != nil {
		query = query.Where("sermon_notes.series_id = ?", *filter.SeriesID)
	}

	if filter.PreacherID != nil {
		query = query.Where("sermon_notes.preacher_id = ?", *filter.PreacherID)
	}

	if filter.StartDate != nil {
		query = query.Where("sermon_notes.sermon_date >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("sermon_notes.sermon_date <= ?", *filter.EndDate)
	}

	if filter.TagID != nil {
		query = query.Joins("JOIN note_tags ON note_tags.note_id = sermon_notes.id").
			Where("note_tags.tag_id = ?", *filter.TagID)
	}

	var notes []*entity.SermonNote

	if err := query.Preload("Verses").Preload("Tags").Find(&notes).Error; err != nil {
		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	for _, note := range notes {
		r.hydrateNoteFile(note)
	}

	r.log.Info("find all completed", zap.Any("param", filter))

	return notes, nil
}

func (r *sermonNoteRepository) Search(
	ctx context.Context,
	query string,
) ([]*entity.SermonNote, error) {
	r.log.Info("search started", zap.Any("param", query))

	var notes []*entity.SermonNote

	pattern := "%" + query + "%"

	err := r.conn(ctx).Table("sermon_notes").
		Joins("JOIN notes_fts ON notes_fts.note_id = sermon_notes.id").
		Where("sermon_notes.deleted_at IS NULL").
		Where("notes_fts.title LIKE ? OR notes_fts.preacher_name LIKE ? OR notes_fts.plain_text LIKE ?", pattern, pattern, pattern).
		Find(&notes).Error

	if err != nil {
		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	r.log.Info("search completed", zap.Any("param", query))

	return notes, nil
}

func (r *sermonNoteRepository) FindByScripture(
	ctx context.Context,
	query entity.ScriptureQuery,
) ([]*entity.SermonNote, error) {
	r.log.Info("find by scripture started", zap.Any("param", query))

	dbQuery := r.conn(ctx).Table("sermon_notes").
		Joins("JOIN note_verses ON note_verses.note_id = sermon_notes.id").
		Where("sermon_notes.deleted_at IS NULL").
		Where("note_verses.translation = ? AND note_verses.book_id = ? AND note_verses.chapter = ? AND note_verses.verse_start >= ?",
			query.Translation, query.BookID, query.Chapter, query.VerseStart)

	if query.VerseEnd != nil {
		dbQuery = dbQuery.Where("note_verses.verse_start <= ?", *query.VerseEnd)
	}

	var notes []*entity.SermonNote

	if err := dbQuery.Distinct("sermon_notes.*").Find(&notes).Error; err != nil {
		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	r.log.Info("find by scripture completed", zap.Any("param", query))

	return notes, nil
}

func (r *sermonNoteRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	sermonNoteModel := &entity.SermonNote{}
	if err := r.conn(ctx).Where("id = ?", id).Delete(sermonNoteModel).Error; err != nil {
		return entityerror.ErrSermonNoteDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}

func (r *sermonNoteRepository) Restore(ctx context.Context, id string) error {
	r.log.Info("restore started", zap.Any("param", id))

	sermonNoteModel := &entity.SermonNote{}
	if err := r.conn(ctx).Unscoped().Model(sermonNoteModel).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	r.log.Info("restore completed", zap.Any("param", id))

	return nil
}

func (r *sermonNoteRepository) FindUpdatedSince(
	ctx context.Context,
	since time.Time,
) ([]*entity.SermonNote, error) {
	r.log.Info("find updated since started", zap.Time("since", since))

	var notes []*entity.SermonNote

	query := r.conn(ctx).Unscoped()
	if !since.IsZero() {
		query = query.Where("updated_at > ?", since)
	}

	err := query.
		Order("updated_at ASC").
		Preload("Verses").
		Preload("Tags").
		Find(&notes).Error
	if err != nil {
		return nil, entityerror.ErrSermonNoteRetrievalFailed(err)
	}

	for _, n := range notes {
		r.hydrateNoteFile(n)
	}

	r.log.Info("find updated since completed", zap.Int("count", len(notes)))

	return notes, nil
}

func (r *sermonNoteRepository) FindAttachmentByID(
	ctx context.Context,
	id string,
) (*entity.SermonNoteAttachment, error) {
	r.log.Info("find attachment by id started", zap.String("id", id))

	if id == "" {
		r.log.Debug("find attachment by id skipped due to empty id")

		return nil, entityerror.ErrSermonNoteAttachmentNotFound(id)
	}

	var attachment entity.SermonNoteAttachment

	err := r.conn(ctx).Where("id = ?", id).First(&attachment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonNoteAttachmentNotFound(id)
		}

		return nil, entityerror.ErrSermonNoteAttachmentRetrievalFailed(err)
	}

	r.log.Info("find attachment by id completed", zap.String("id", id))

	return &attachment, nil
}

func (r *sermonNoteRepository) UpdateAttachment(
	ctx context.Context,
	attachment *entity.SermonNoteAttachment,
) error {
	r.log.Info("update attachment started")

	if attachment == nil {
		err := errors.New("attachment is nil")

		return entityerror.ErrSermonNoteAttachmentUpdateFailed(err)
	}

	updates := map[string]any{
		"storage_path": attachment.StoragePath,
		"status":       attachment.Status,
		"file_size":    attachment.FileSize,
		"updated_at":   attachment.UpdatedAt,
	}

	err := r.conn(ctx).
		Model(&entity.SermonNoteAttachment{}).
		Where("id = ?", attachment.ID).
		Updates(updates).Error
	if err != nil {
		return entityerror.ErrSermonNoteAttachmentUpdateFailed(err)
	}

	r.log.Info("update attachment completed", zap.String("id", attachment.ID))

	return nil
}
