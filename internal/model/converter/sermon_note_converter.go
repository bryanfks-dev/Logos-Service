package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func SyncNoteChangesItemToEntity(
	noteItem *request.SyncNoteChangesItem,
) entity.SermonNote {
	if noteItem == nil {
		return entity.SermonNote{}
	}

	version := noteItem.Version
	if version < 1 {
		version = 1
	}

	verses := make([]entity.NoteVerse, 0, len(noteItem.Verses))
	for _, v := range noteItem.Verses {
		verse := entity.NoteVerse{
			ID:          v.ID,
			NoteID:      noteItem.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		}
		verses = append(verses, verse)
	}

	tags := make([]entity.Tag, 0, len(noteItem.Tags))
	for _, tagName := range noteItem.Tags {
		tag := entity.Tag{
			Name: tagName,
		}
		tags = append(tags, tag)
	}

	sermonDate := time.Now()
	if noteItem.SermonDate != nil {
		sermonDate = *noteItem.SermonDate
	}

	return entity.SermonNote{
		ID:          noteItem.ID,
		GatheringID: noteItem.GatheringID,
		SeriesID:    noteItem.SeriesID,
		PreacherID:  noteItem.PreacherID,
		Title:       noteItem.Title,
		SermonDate:  sermonDate,
		ContentHash: noteItem.ContentHash,
		PlainText:   noteItem.PlainText,
		ContentJSON: noteItem.ContentJSON,
		Version:     version,
		Verses:      verses,
		Tags:        tags,
	}
}

func SermonNoteToResponse(
	note *entity.SermonNote,
) *response.SermonNoteResponse {
	if note == nil {
		return nil
	}

	tags := make([]string, 0, len(note.Tags))
	for _, tag := range note.Tags {
		tags = append(tags, tag.Name)
	}

	verses := make([]response.SermonNoteVerseResponse, 0, len(note.Verses))
	for _, v := range note.Verses {
		verse := response.SermonNoteVerseResponse{
			ID:          v.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		}
		verses = append(verses, verse)
	}

	var deletedAt *time.Time
	if note.DeletedAt.Valid {
		deletedAt = &note.DeletedAt.Time
	}

	return &response.SermonNoteResponse{
		ID:          note.ID,
		GatheringID: note.GatheringID,
		SeriesID:    note.SeriesID,
		PreacherID:  note.PreacherID,
		Title:       note.Title,
		SermonDate:  note.SermonDate,
		FilePath:    note.FilePath,
		ContentHash: note.ContentHash,
		PlainText:   note.PlainText,
		ContentJSON: note.ContentJSON,
		Version:     note.Version,
		SyncStatus:  note.SyncStatus,
		Tags:        tags,
		Verses:      verses,
		CreatedAt:   note.CreatedAt,
		UpdatedAt:   note.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}
