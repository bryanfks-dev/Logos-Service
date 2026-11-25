package model_test

import (
	"strings"
	"testing"

	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncPushRequest_Validate(t *testing.T) {
	t.Run("accepts empty sync push request", func(t *testing.T) {
		req := model.SyncPushRequest{}
		assert.NoError(t, req.Validate())
	})

	t.Run("accepts valid nested items across all entities", func(t *testing.T) {
		verseEnd := 5
		req := model.SyncPushRequest{
			Gatherings: []model.SyncGatheringItem{
				{
					ID:        "g-1",
					Name:      "Sunday Service",
					ColorHex:  "#112233",
					SortOrder: 1,
				},
			},
			Series: []model.SyncSeriesItem{
				{ID: "s-1", Name: "Romans"},
			},
			Preachers: []model.SyncPreacherItem{
				{ID: "p-1", Name: "Pastor John", ChurchName: "Grace Church"},
			},
			Tags: []model.SyncTagItem{
				{ID: "t-1", Name: "Grace"},
			},
			Notes: []model.SyncNoteItem{
				{
					ID:          "n-1",
					Title:       "Grace Alone",
					ContentHash: "abc123",
					Version:     1,
					Tags:        []string{"Grace"},
					Verses: []model.SyncNoteVerse{
						{
							ID:          "v-1",
							Translation: "ESV",
							BookID:      45,
							Chapter:     3,
							VerseStart:  1,
							VerseEnd:    &verseEnd,
							DisplayRef:  "Romans 3:1-5",
						},
					},
				},
			},
		}

		assert.NoError(t, req.Validate())
	})

	t.Run("rejects invalid nested items", func(t *testing.T) {
		req := model.SyncPushRequest{
			Gatherings: []model.SyncGatheringItem{
				{Name: "", ColorHex: "not-a-hex", SortOrder: -1},
			},
			Series: []model.SyncSeriesItem{
				{ID: strings.Repeat("a", 37), Name: ""},
			},
			Preachers: []model.SyncPreacherItem{
				{Name: ""},
			},
			Tags: []model.SyncTagItem{
				{Name: strings.Repeat("x", 101)},
			},
			Notes: []model.SyncNoteItem{
				{
					Title:   "",
					Version: -1,
					Tags:    []string{""},
					Verses: []model.SyncNoteVerse{
						{Translation: "", BookID: 0, Chapter: 0, VerseStart: 0},
					},
				},
			},
		}

		err := req.Validate()
		require.Error(t, err)
	})
}

func TestSyncNotesRequest_Validate(t *testing.T) {
	t.Run("requires notes", func(t *testing.T) {
		req := model.SyncNotesRequest{}
		require.Error(t, req.Validate())
	})

	t.Run("accepts valid notes", func(t *testing.T) {
		req := model.SyncNotesRequest{
			Notes: []model.SyncNoteItem{
				{ID: "n-1", Title: "Faith", Version: 1},
			},
		}
		assert.NoError(t, req.Validate())
	})
}
