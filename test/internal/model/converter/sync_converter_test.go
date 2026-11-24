package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/model/converter"
	"github.com/bryanfks-dev/selah/pkg/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNoteEntityToSyncItem(t *testing.T) {
	t.Run("nil note returns empty item", func(t *testing.T) {
		item := converter.NoteEntityToSyncItem(nil)
		assert.Empty(t, item.ID)
	})

	t.Run("converts full note entity to sync item", func(t *testing.T) {
		now := time.Now()
		vEnd := 5
		note := &entity.SermonNote{
			ID:          "note-123",
			Title:       "Grace and Peace",
			PlainText:   "Text content",
			ContentJSON: `{"doc": true}`,
			Version:     2,
			UpdatedAt:   now,
			DeletedAt:   gorm.DeletedAt{Time: now, Valid: true},
			Tags: []entity.Tag{
				{Name: "Grace"},
				{Name: "Faith"},
			},
			Verses: []entity.NoteVerse{
				{
					ID:          "verse-1",
					Translation: "ESV",
					BookID:      1,
					Chapter:     1,
					VerseStart:  1,
					VerseEnd:    &vEnd,
					DisplayRef:  "Genesis 1:1-5",
				},
			},
		}

		item := converter.NoteEntityToSyncItem(note)

		assert.Equal(t, "note-123", item.ID)
		assert.Equal(t, "Grace and Peace", item.Title)
		assert.Equal(t, 2, item.Version)
		assert.True(t, item.IsDeleted)
		assert.ElementsMatch(t, []string{"Grace", "Faith"}, item.Tags)
		assert.Len(t, item.Verses, 1)
		assert.Equal(t, "ESV", item.Verses[0].Translation)
		assert.Equal(t, 1, item.Verses[0].BookID)
		assert.Equal(t, &vEnd, item.Verses[0].VerseEnd)
	})
}

func TestSyncItemToNoteEntity(t *testing.T) {
	t.Run("nil item returns empty entity", func(t *testing.T) {
		note := converter.SyncItemToNoteEntity(nil)
		assert.Empty(t, note.ID)
	})

	t.Run("converts sync item to note entity", func(t *testing.T) {
		item := &model.SyncNoteItem{
			ID:          "item-123",
			Title:       "Faith",
			PlainText:   "Faith content",
			ContentJSON: `{"root": true}`,
			Version:     0, // should default to 1
			Tags:        []string{"Hope", "Love"},
			Verses: []model.SyncNoteVerse{
				{
					ID:          "v-1",
					Translation: "NIV",
					BookID:      40,
					Chapter:     5,
					VerseStart:  1,
					DisplayRef:  "Matthew 5:1",
				},
			},
		}

		note := converter.SyncItemToNoteEntity(item)

		assert.Equal(t, "item-123", note.ID)
		assert.Equal(t, "Faith", note.Title)
		assert.Equal(t, 1, note.Version)
		assert.Len(t, note.Tags, 2)
		assert.Equal(t, "Hope", note.Tags[0].Name)
		assert.Len(t, note.Verses, 1)
		assert.Equal(t, "item-123", note.Verses[0].NoteID)
		assert.Equal(t, "NIV", note.Verses[0].Translation)
	})
}
