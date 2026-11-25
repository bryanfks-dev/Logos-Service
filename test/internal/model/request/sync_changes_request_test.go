package request_test

import (
	"strings"
	"testing"

	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncPushChangesRequest_Validate(t *testing.T) {
	t.Run("accepts empty sync push request", func(t *testing.T) {
		req := request.SyncPushChangesRequest{}
		assert.NoError(t, req.Validate())
	})

	t.Run("accepts valid nested items across all entities", func(t *testing.T) {
		verseEnd := 5
		req := request.SyncPushChangesRequest{
			Gatherings: []request.SyncGatheringChangesItem{
				{
					ID:        "g-1",
					Name:      "Sunday Service",
					ColorHex:  "#112233",
					SortOrder: 1,
				},
			},
			Series: []request.SyncSeriesChangesItem{
				{ID: "s-1", Name: "Romans"},
			},
			Preachers: []request.SyncPreacherChangesItem{
				{ID: "p-1", Name: "Pastor John", ChurchName: "Grace Church"},
			},
			Tags: []request.SyncTagChangesItem{
				{ID: "t-1", Name: "Grace"},
			},
			SermonNotes: []request.SyncNoteChangesItem{
				{
					ID:          "n-1",
					Title:       "Grace Alone",
					ContentHash: "abc123",
					Version:     1,
					Tags:        []string{"Grace"},
					Verses: []request.SyncNoteVerseChangesItem{
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
		req := request.SyncPushChangesRequest{
			Gatherings: []request.SyncGatheringChangesItem{
				{Name: "", ColorHex: "not-a-hex", SortOrder: -1},
			},
			Series: []request.SyncSeriesChangesItem{
				{ID: strings.Repeat("a", 37), Name: ""},
			},
			Preachers: []request.SyncPreacherChangesItem{
				{Name: ""},
			},
			Tags: []request.SyncTagChangesItem{
				{Name: strings.Repeat("x", 101)},
			},
			SermonNotes: []request.SyncNoteChangesItem{
				{
					Title:   "",
					Version: -1,
					Tags:    []string{""},
					Verses: []request.SyncNoteVerseChangesItem{
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
	t.Run("accepts empty sync push changes request", func(t *testing.T) {
		req := request.SyncPushChangesRequest{}
		assert.NoError(t, req.Validate())
	})

	t.Run("accepts valid notes", func(t *testing.T) {
		req := request.SyncPushChangesRequest{
			SermonNotes: []request.SyncNoteChangesItem{
				{ID: "n-1", Title: "Faith", Version: 1},
			},
		}
		assert.NoError(t, req.Validate())
	})
}
