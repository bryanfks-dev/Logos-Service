package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	dbcorrelation "github.com/bryanfks-dev/selah/infrastructure/db/correlation"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SermonNoteRepository interface {
	Create(ctx context.Context, note *entity.SermonNote) error
	Update(ctx context.Context, note *entity.SermonNote) error
	FindByID(ctx context.Context, id string) (*entity.SermonNote, error)
	FindByIDUnscoped(ctx context.Context, id string) (*entity.SermonNote, error)
	FindAll(
		ctx context.Context,
		filter entity.SermonNoteFilter,
	) ([]entity.SermonNote, error)
	Search(ctx context.Context, query string) ([]entity.SermonNote, error)
	FindByScripture(
		ctx context.Context,
		query entity.ScriptureQuery,
	) ([]entity.SermonNote, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
}

type sermonNoteRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSermonNoteRepository(
	db *gorm.DB,
	log *zap.Logger,
) SermonNoteRepository {
	return &sermonNoteRepository{
		db:  db,
		log: log,
	}
}

func (r *sermonNoteRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := dbcorrelation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *sermonNoteRepository) persistNoteFile(note *entity.SermonNote) error {
	if note.ContentJSON == "" {
		return nil
	}

	if note.FilePath == "" {
		dir := filepath.Join(os.TempDir(), "selah_notes")

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		note.FilePath = filepath.Join(dir, note.ID+".json")
	} else {
		dir := filepath.Dir(note.FilePath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(note.FilePath, []byte(note.ContentJSON), 0o644)
}

func (r *sermonNoteRepository) hydrateNoteFile(note *entity.SermonNote) {
	if note.FilePath != "" {
		if data, err := os.ReadFile(note.FilePath); err == nil {
			note.ContentJSON = string(data)
		}
	}
}

func (r *sermonNoteRepository) saveChildren(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	if err := r.getDB(ctx).Where("note_id = ?", note.ID).Delete(&entity.NoteVerse{}).Error; err != nil {
		return err
	}

	for i := range note.Verses {
		note.Verses[i].NoteID = note.ID

		if err := r.getDB(ctx).Create(&note.Verses[i]).Error; err != nil {
			return err
		}
	}

	if err := r.getDB(ctx).Where("note_id = ?", note.ID).Delete(&entity.NoteTag{}).Error; err != nil {
		return err
	}

	for _, tag := range note.Tags {
		noteTag := entity.NoteTag{
			NoteID: note.ID,
			TagID:  tag.ID,
		}

		if err := r.getDB(ctx).Create(&noteTag).Error; err != nil {
			return err
		}
	}

	if err := r.getDB(ctx).Where("note_id = ?", note.ID).Delete(&entity.NotesFTS{}).Error; err != nil {
		return err
	}

	if note.FTS != nil {
		note.FTS.NoteID = note.ID

		if err := r.getDB(ctx).Create(note.FTS).Error; err != nil {
			return err
		}
	} else {
		fts := &entity.NotesFTS{
			NoteID:    note.ID,
			Title:     note.Title,
			PlainText: note.PlainText,
		}

		if err := r.getDB(ctx).Create(fts).Error; err != nil {
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

	db := r.getDB(ctx)

	if err := db.Create(note).Error; err != nil {
		r.log.Error("create failed", zap.Error(err), zap.Any("param", note))

		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	if err := r.persistNoteFile(note); err != nil {
		r.log.Error("create failed", zap.Error(err), zap.Any("param", note))

		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	if err := r.saveChildren(ctx, note); err != nil {
		r.log.Error("create failed", zap.Error(err), zap.Any("param", note))

		return entityerror.ErrSermonNoteCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", note))

	return nil
}

func (r *sermonNoteRepository) Update(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	r.log.Info("update started", zap.Any("param", note))

	if err := r.persistNoteFile(note); err != nil {
		r.log.Error("update failed", zap.Error(err), zap.Any("param", note))

		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	db := r.getDB(ctx)

	if err := db.Save(note).Error; err != nil {
		r.log.Error("update failed", zap.Error(err), zap.Any("param", note))

		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	if err := r.saveChildren(ctx, note); err != nil {
		r.log.Error("update failed", zap.Error(err), zap.Any("param", note))

		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", note))

	return nil
}

func (r *sermonNoteRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.SermonNote, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var note entity.SermonNote

	if err := r.getDB(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&note).Error; err != nil {
		r.log.Error("find by id failed", zap.Error(err), zap.Any("param", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonNoteNotFound(id)
		}

		return nil, entityerror.ErrSermonNoteFindByIDFailed(err)
	}

	var verses []entity.NoteVerse
	_ = r.getDB(ctx).Where("note_id = ?", id).Find(&verses).Error
	note.Verses = verses

	var tags []entity.Tag
	_ = r.getDB(ctx).Table("tags").
		Joins("JOIN note_tags ON note_tags.tag_id = tags.id").
		Where("note_tags.note_id = ?", id).
		Find(&tags).
		Error
	note.Tags = tags

	var fts entity.NotesFTS
	if err := r.getDB(ctx).Where("note_id = ?", id).First(&fts).Error; err == nil {
		note.FTS = &fts
	}

	r.hydrateNoteFile(&note)

	r.log.Info("find by id completed", zap.Any("param", id))

	return &note, nil
}

func (r *sermonNoteRepository) FindByIDUnscoped(
	ctx context.Context,
	id string,
) (*entity.SermonNote, error) {
	r.log.Info("find by id unscoped started", zap.Any("param", id))

	var note entity.SermonNote

	if err := r.getDB(ctx).Unscoped().Where("id = ?", id).First(&note).Error; err != nil {
		r.log.Error(
			"find by id unscoped failed",
			zap.Error(err),
			zap.Any("param", id),
		)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrSermonNoteNotFound(id)
		}

		return nil, entityerror.ErrSermonNoteFindByIDFailed(err)
	}

	var verses []entity.NoteVerse
	_ = r.getDB(ctx).Where("note_id = ?", id).Find(&verses).Error
	note.Verses = verses

	var tags []entity.Tag
	_ = r.getDB(ctx).Table("tags").
		Joins("JOIN note_tags ON note_tags.tag_id = tags.id").
		Where("note_tags.note_id = ?", id).
		Find(&tags).
		Error
	note.Tags = tags

	var fts entity.NotesFTS
	if err := r.getDB(ctx).Where("note_id = ?", id).First(&fts).Error; err == nil {
		note.FTS = &fts
	}

	r.hydrateNoteFile(&note)

	r.log.Info("find by id unscoped completed", zap.Any("param", id))

	return &note, nil
}

func (r *sermonNoteRepository) FindAll(
	ctx context.Context,
	filter entity.SermonNoteFilter,
) ([]entity.SermonNote, error) {
	r.log.Info("find all started", zap.Any("param", filter))

	db := r.getDB(ctx)
	if filter.IncludeDeleted {
		db = db.Unscoped()
	}

	query := db.Model(&entity.SermonNote{})

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

	var notes []entity.SermonNote

	if err := query.Preload("Verses").Preload("Tags").Find(&notes).Error; err != nil {
		r.log.Error("find all failed", zap.Error(err), zap.Any("param", filter))

		return nil, entityerror.ErrSermonNoteFindAllFailed(err)
	}

	for i := range notes {
		r.hydrateNoteFile(&notes[i])
	}

	r.log.Info("find all completed", zap.Any("param", filter))

	return notes, nil
}

func (r *sermonNoteRepository) Search(
	ctx context.Context,
	query string,
) ([]entity.SermonNote, error) {
	r.log.Info("search started", zap.Any("param", query))

	var notes []entity.SermonNote

	pattern := "%" + query + "%"

	err := r.getDB(ctx).Table("sermon_notes").
		Joins("JOIN notes_fts ON notes_fts.note_id = sermon_notes.id").
		Where("sermon_notes.deleted_at IS NULL").
		Where("notes_fts.title LIKE ? OR notes_fts.preacher_name LIKE ? OR notes_fts.plain_text LIKE ?", pattern, pattern, pattern).
		Find(&notes).Error

	if err != nil {
		r.log.Error("search failed", zap.Error(err), zap.Any("param", query))

		return nil, entityerror.ErrSermonNoteSearchFailed(err)
	}

	r.log.Info("search completed", zap.Any("param", query))

	return notes, nil
}

func (r *sermonNoteRepository) FindByScripture(
	ctx context.Context,
	query entity.ScriptureQuery,
) ([]entity.SermonNote, error) {
	r.log.Info("find by scripture started", zap.Any("param", query))

	dbQuery := r.getDB(ctx).Table("sermon_notes").
		Joins("JOIN note_verses ON note_verses.note_id = sermon_notes.id").
		Where("sermon_notes.deleted_at IS NULL").
		Where("note_verses.translation = ? AND note_verses.book_id = ? AND note_verses.chapter = ? AND note_verses.verse_start >= ?",
			query.Translation, query.BookID, query.Chapter, query.VerseStart)

	if query.VerseEnd != nil {
		dbQuery = dbQuery.Where("note_verses.verse_start <= ?", *query.VerseEnd)
	}

	var notes []entity.SermonNote

	if err := dbQuery.Distinct("sermon_notes.*").Find(&notes).Error; err != nil {
		r.log.Error(
			"find by scripture failed",
			zap.Error(err),
			zap.Any("param", query),
		)

		return nil, entityerror.ErrSermonNoteFindByScriptureFailed(err)
	}

	r.log.Info("find by scripture completed", zap.Any("param", query))

	return notes, nil
}

func (r *sermonNoteRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	if err := r.getDB(ctx).Where("id = ?", id).Delete(&entity.SermonNote{}).Error; err != nil {
		r.log.Error("delete failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrSermonNoteDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}

func (r *sermonNoteRepository) Restore(ctx context.Context, id string) error {
	r.log.Info("restore started", zap.Any("param", id))

	if err := r.getDB(ctx).Unscoped().Model(&entity.SermonNote{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		r.log.Error("restore failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrSermonNoteUpdateFailed(err)
	}

	r.log.Info("restore completed", zap.Any("param", id))

	return nil
}
