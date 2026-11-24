package converter

import (
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/pkg/entity"
)

func NoteEntityToSyncItem(note *entity.SermonNote) model.SyncNoteItem {
	if note == nil {
		return model.SyncNoteItem{}
	}

	tags := make([]string, 0, len(note.Tags))
	for _, tag := range note.Tags {
		tags = append(tags, tag.Name)
	}

	verses := make([]model.SyncNoteVerse, 0, len(note.Verses))
	for _, v := range note.Verses {
		verses = append(verses, model.SyncNoteVerse{
			ID:          v.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		})
	}

	isDeleted := note.DeletedAt.Valid
	updatedAt := note.UpdatedAt

	return model.SyncNoteItem{
		ID:          note.ID,
		GatheringID: note.GatheringID,
		SeriesID:    note.SeriesID,
		PreacherID:  note.PreacherID,
		Title:       note.Title,
		SermonDate:  note.SermonDate,
		PlainText:   note.PlainText,
		ContentJSON: note.ContentJSON,
		Version:     note.Version,
		Tags:        tags,
		Verses:      verses,
		IsDeleted:   isDeleted,
		UpdatedAt:   &updatedAt,
	}
}

func SyncItemToNoteEntity(item *model.SyncNoteItem) entity.SermonNote {
	if item == nil {
		return entity.SermonNote{}
	}

	version := item.Version
	if version < 1 {
		version = 1
	}

	verses := make([]entity.NoteVerse, 0, len(item.Verses))
	for _, v := range item.Verses {
		verses = append(verses, entity.NoteVerse{
			ID:          v.ID,
			NoteID:      item.ID,
			Translation: v.Translation,
			BookID:      v.BookID,
			Chapter:     v.Chapter,
			VerseStart:  v.VerseStart,
			VerseEnd:    v.VerseEnd,
			DisplayRef:  v.DisplayRef,
		})
	}

	tags := make([]entity.Tag, 0, len(item.Tags))
	for _, tagName := range item.Tags {
		tags = append(tags, entity.Tag{
			Name: tagName,
		})
	}

	return entity.SermonNote{
		ID:          item.ID,
		GatheringID: item.GatheringID,
		SeriesID:    item.SeriesID,
		PreacherID:  item.PreacherID,
		Title:       item.Title,
		SermonDate:  item.SermonDate,
		PlainText:   item.PlainText,
		ContentJSON: item.ContentJSON,
		Version:     version,
		Verses:      verses,
		Tags:        tags,
	}
}
