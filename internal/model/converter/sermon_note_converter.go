package converter

import (
	"time"

	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/internal/model/response"
	pushsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/push_sync_changes"
	"github.com/bryanfks-dev/logos/pkg/entity"
)

func SyncNoteChangesItemToPushSermonNoteInput(
	noteItem *request.SyncNoteChangesItem,
) *pushsyncchanges.PushSermonNoteInput {
	if noteItem == nil {
		return nil
	}

	verses := make(
		[]pushsyncchanges.PushNoteVerseInput,
		0,
		len(noteItem.Verses),
	)
	for _, v := range noteItem.Verses {
		verse := pushsyncchanges.PushNoteVerseInput{
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

	tags := make([]pushsyncchanges.PushNoteTagInput, 0, len(noteItem.Tags))
	for _, tagName := range noteItem.Tags {
		tag := pushsyncchanges.PushNoteTagInput{
			Name: tagName,
		}
		tags = append(tags, tag)
	}

	var sermonDate time.Time
	if noteItem.SermonDate != nil {
		sermonDate = *noteItem.SermonDate
	}

	var createdAt time.Time
	if noteItem.CreateadAt != nil {
		createdAt = *noteItem.CreateadAt
	}

	var updatedAt time.Time
	if noteItem.UpdatedAt != nil {
		updatedAt = *noteItem.UpdatedAt
	}

	return &pushsyncchanges.PushSermonNoteInput{
		ID:          noteItem.ID,
		GatheringID: noteItem.GatheringID,
		SeriesID:    noteItem.SeriesID,
		PreacherID:  noteItem.PreacherID,
		Title:       noteItem.Title,
		SermonDate:  sermonDate,
		ContentHash: noteItem.ContentHash,
		PlainText:   noteItem.PlainText,
		ContentJSON: noteItem.ContentJSON,
		Version:     noteItem.Version,
		Tags:        tags,
		Verses:      verses,
		IsDeleted:   noteItem.IsDeleted,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
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
