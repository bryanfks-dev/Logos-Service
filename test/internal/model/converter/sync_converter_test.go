package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSermonNoteToSyncItem(t *testing.T) {
	t.Run("nil note returns empty item", func(t *testing.T) {
		item := converter.SermonNoteToSyncItem(nil)
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

		item := converter.SermonNoteToSyncItem(note)

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

func TestSyncItemToSermonNote(t *testing.T) {
	t.Run("nil item returns empty entity", func(t *testing.T) {
		note := converter.SyncItemToSermonNote(nil)
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
			Verses: []request.SyncNoteVerse{
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

		note := converter.SyncItemToSermonNote(noteItem)

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

func TestSyncPushRequestToBatch(t *testing.T) {
	t.Run("nil request returns empty batch", func(t *testing.T) {
		batch := converter.SyncPushChangesRequestToBatch(nil)
		assert.NotNil(t, batch)
		assert.Empty(t, batch.Notes)
		assert.Empty(t, batch.Gatherings)
		assert.Empty(t, batch.Series)
		assert.Empty(t, batch.Preachers)
		assert.Empty(t, batch.Tags)
	})

	t.Run("converts full sync push request to batch", func(t *testing.T) {
		now := time.Now()
		req := &request.SyncPushChangesRequest{
			SermonNotes: []request.SyncNoteChangesItem{
				{
					ID:          "note-1",
					Title:       "Test Note",
					ContentHash: "hash-abc",
					Version:     1,
					UpdatedAt:   &now,
				},
			},
			Gatherings: []request.SyncGatheringChangesItem{
				{
					ID:        "g-1",
					Name:      "Sunday Morning",
					ColorHex:  "#FFFFFF",
					SortOrder: 1,
					UpdatedAt: &now,
				},
			},
			Series: []request.SyncSeriesChangesItem{
				{
					ID:          "s-1",
					Name:        "Romans",
					Description: "Study on Romans",
					UpdatedAt:   &now,
				},
			},
			Preachers: []request.SyncPreacherChangesItem{
				{
					ID:         "p-1",
					Name:       "Pastor John",
					ChurchName: "Grace Church",
					UpdatedAt:  &now,
				},
			},
			Tags: []request.SyncTagChangesItem{
				{
					ID:        "t-1",
					Name:      "Theology",
					UpdatedAt: &now,
				},
			},
		}

		batch := converter.SyncPushChangesRequestToBatch(req)
		assert.NotNil(t, batch)

		require.Len(t, batch.Notes, 1)
		assert.Equal(t, "note-1", batch.Notes[0].ID)
		assert.Equal(t, "Test Note", batch.Notes[0].Title)
		assert.Equal(t, "hash-abc", batch.Notes[0].ContentHash)

		require.Len(t, batch.Gatherings, 1)
		assert.Equal(t, "g-1", batch.Gatherings[0].ID)
		assert.Equal(t, "Sunday Morning", batch.Gatherings[0].Name)
		assert.Equal(t, "#FFFFFF", batch.Gatherings[0].ColorHex)

		require.Len(t, batch.Series, 1)
		assert.Equal(t, "s-1", batch.Series[0].ID)
		assert.Equal(t, "Romans", batch.Series[0].Name)

		require.Len(t, batch.Preachers, 1)
		assert.Equal(t, "p-1", batch.Preachers[0].ID)
		assert.Equal(t, "Pastor John", batch.Preachers[0].Name)

		require.Len(t, batch.Tags, 1)
		assert.Equal(t, "t-1", batch.Tags[0].ID)
		assert.Equal(t, "Theology", batch.Tags[0].Name)
	})
}

func TestSyncPushResultToResponse(t *testing.T) {
	t.Run(
		"nil result returns empty response with non-nil slices",
		func(t *testing.T) {
			res := converter.SyncPushResultToResponse(nil)
			assert.NotNil(t, res)
			assert.True(t, res.SyncedAt.IsZero())
			assert.NotNil(t, res.MissingContentIDs)
			assert.Empty(t, res.MissingContentIDs)
			assert.NotNil(t, res.Conflicts)
			assert.Empty(t, res.Conflicts)
		},
	)

	t.Run("converts result to response", func(t *testing.T) {
		syncedAt := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
		result := &model.SyncPushChangesResult{
			SyncedAt:          syncedAt,
			MissingContentIDs: []string{"n1", "n2"},
			Conflicts:         []string{"n3"},
		}

		res := converter.SyncPushResultToResponse(result)
		assert.NotNil(t, res)
		assert.Equal(t, syncedAt, res.SyncedAt)
		assert.Equal(t, []string{"n1", "n2"}, res.MissingContentIDs)
		assert.Equal(t, []string{"n3"}, res.Conflicts)
	})
}

func TestHelperConverters(t *testing.T) {
	now := time.Now()

	t.Run("nil converters return nil", func(t *testing.T) {
		assert.Nil(t, converter.SermonNoteToResponse(nil))
		assert.Nil(t, converter.GatheringToResponse(nil))
		assert.Nil(t, converter.SermonSeriesToResponse(nil))
		assert.Nil(t, converter.PreacherToResponse(nil))
		assert.Nil(t, converter.TagToResponse(nil))
	})

	t.Run(
		"SermonNoteToResponse converts active and soft-deleted notes",
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
			assert.False(t, res.IsDeleted)
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
			assert.True(t, resDel.IsDeleted)
			require.NotNil(t, resDel.DeletedAt)
			assert.Equal(t, now, *resDel.DeletedAt)
		},
	)

	t.Run("GatheringToResponse converts gathering", func(t *testing.T) {
		g := &entity.Gathering{
			ID:        "g-1",
			Name:      "Sunday",
			ColorHex:  "#111111",
			SortOrder: 2,
			CreatedAt: now,
			UpdatedAt: now,
		}
		res := converter.GatheringToResponse(g)
		require.NotNil(t, res)
		assert.Equal(t, "g-1", res.ID)
		assert.Equal(t, "Sunday", res.Name)
		assert.Equal(t, "#111111", res.ColorHex)
		assert.Equal(t, 2, res.SortOrder)
		assert.Nil(t, res.DeletedAt)
	})

	t.Run("SermonSeriesToResponse converts series", func(t *testing.T) {
		s := &entity.SermonSeries{
			ID:          "s-1",
			Name:        "Genesis",
			Description: "Origins",
			CreatedAt:   now,
			UpdatedAt:   now,
			DeletedAt:   gorm.DeletedAt{Time: now, Valid: true},
		}
		res := converter.SermonSeriesToResponse(s)
		require.NotNil(t, res)
		assert.Equal(t, "s-1", res.ID)
		assert.Equal(t, "Genesis", res.Name)
		assert.NotNil(t, res.DeletedAt)
	})

	t.Run("PreacherToResponse converts preacher", func(t *testing.T) {
		p := &entity.Preacher{
			ID:         "p-1",
			Name:       "John",
			ChurchName: "Church",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		res := converter.PreacherToResponse(p)
		require.NotNil(t, res)
		assert.Equal(t, "p-1", res.ID)
		assert.Equal(t, "John", res.Name)
		assert.Nil(t, res.DeletedAt)
	})

	t.Run("TagToResponse converts tag", func(t *testing.T) {
		tag := &entity.Tag{
			ID:         "t-1",
			Name:       "Grace",
			UsageCount: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		res := converter.TagToResponse(tag)
		require.NotNil(t, res)
		assert.Equal(t, "t-1", res.ID)
		assert.Equal(t, "Grace", res.Name)
		assert.Equal(t, 5, res.UsageCount)
	})
}

func TestSyncPullChangesResultToResponse(t *testing.T) {
	now := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)

	t.Run(
		"nil result returns empty response with non-nil slices",
		func(t *testing.T) {
			res := converter.SyncPullChangesResultToResponse(nil)
			require.NotNil(t, res)
			assert.True(t, res.ServerTime.IsZero())
			assert.Empty(t, res.Notes)
			assert.Empty(t, res.Gatherings)
			assert.Empty(t, res.Series)
			assert.Empty(t, res.Preachers)
			assert.Empty(t, res.Tags)
		},
	)

	t.Run(
		"converts full pull result with datetime server time",
		func(t *testing.T) {
			result := &model.SyncPullChangesResult{
				ServerTime: now,
				Notes: []*entity.SermonNote{
					{
						ID:        "n-1",
						Title:     "Note 1",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				Gatherings: []*entity.Gathering{
					{
						ID:        "g-1",
						Name:      "Gathering 1",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				Series: []*entity.SermonSeries{
					{
						ID:        "s-1",
						Name:      "Series 1",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				Preachers: []*entity.Preacher{
					{
						ID:        "p-1",
						Name:      "Preacher 1",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				Tags: []*entity.Tag{
					{
						ID:        "t-1",
						Name:      "Tag 1",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
			}

			res := converter.SyncPullChangesResultToResponse(result)
			require.NotNil(t, res)
			assert.Equal(t, now, res.ServerTime)
			require.Len(t, res.Notes, 1)
			assert.Equal(t, "n-1", res.Notes[0].ID)
			require.Len(t, res.Gatherings, 1)
			assert.Equal(t, "g-1", res.Gatherings[0].ID)
			require.Len(t, res.Series, 1)
			assert.Equal(t, "s-1", res.Series[0].ID)
			require.Len(t, res.Preachers, 1)
			assert.Equal(t, "p-1", res.Preachers[0].ID)
			require.Len(t, res.Tags, 1)
			assert.Equal(t, "t-1", res.Tags[0].ID)
		},
	)
}
