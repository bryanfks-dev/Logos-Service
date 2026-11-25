package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSyncItemToSermonNote(t *testing.T) {
	t.Run("nil item returns empty entity", func(t *testing.T) {
		note := converter.SyncNoteChangesItemToEntity(nil)
		assert.Empty(t, note.ID)
	})

	t.Run("converts sync item to note entity", func(t *testing.T) {
		noteItem := &request.SyncNoteChangesItem{
			ID:          "item-123",
			Title:       "Faith",
			PlainText:   "Faith content",
			ContentJSON: `{"root": true}`,
			Version:     0, // should default to 1
			Tags:        []string{"Hope", "Love"},
			Verses: []request.SyncNoteVerseChangesItem{
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

		note := converter.SyncNoteChangesItemToEntity(noteItem)

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

func TestSermonNoteToResponse(t *testing.T) {
	now := time.Now()

	t.Run("nil note returns nil", func(t *testing.T) {
		assert.Nil(t, converter.SermonNoteToResponse(nil))
	})

	t.Run(
		"converts active and soft-deleted notes",
		func(t *testing.T) {
			activeNote := &entity.SermonNote{
				ID:          "note-1",
				Title:       "Grace",
				PlainText:   "Text",
				ContentHash: "hash-1",
				Version:     1,
				CreatedAt:   now,
				UpdatedAt:   now,
				Tags:        []entity.Tag{{Name: "Faith"}},
				Verses: []entity.NoteVerse{
					{
						ID:          "v-1",
						Translation: "ESV",
						BookID:      1,
						Chapter:     1,
						VerseStart:  1,
						DisplayRef:  "Gen 1:1",
					},
				},
			}

			res := converter.SermonNoteToResponse(activeNote)
			require.NotNil(t, res)
			assert.Equal(t, "note-1", res.ID)
			assert.Equal(t, "Grace", res.Title)
			assert.Equal(t, "hash-1", res.ContentHash)
			assert.Nil(t, res.DeletedAt)
			assert.Equal(t, []string{"Faith"}, res.Tags)
			require.Len(t, res.Verses, 1)

			deletedNote := &entity.SermonNote{
				ID:        "note-del",
				Title:     "Deleted",
				DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
			}
			resDel := converter.SermonNoteToResponse(deletedNote)
			require.NotNil(t, resDel)
			require.NotNil(t, resDel.DeletedAt)
			assert.Equal(t, now, *resDel.DeletedAt)
		},
	)
}
