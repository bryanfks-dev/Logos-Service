package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/internal/model/request"
	pullsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/pull_sync_changes"
	pushsyncchanges "github.com/bryanfks-dev/logos/internal/usecase/push_sync_changes"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncPushChangesRequestToPushSyncChangesInput(t *testing.T) {
	t.Run("nil request returns empty input", func(t *testing.T) {
		input := converter.SyncPushChangesRequestToPushSyncChangesInput(nil)
		assert.NotNil(t, input)
		assert.Empty(t, input.SermonNotes)
		assert.Empty(t, input.Gatherings)
		assert.Empty(t, input.SermonSeries)
		assert.Empty(t, input.Preachers)
		assert.Empty(t, input.Tags)
	})

	t.Run("converts full sync push request to input", func(t *testing.T) {
		now := time.Now().UTC()
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
			SermonSeries: []request.SyncSeriesChangesItem{
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

		input := converter.SyncPushChangesRequestToPushSyncChangesInput(req)
		assert.NotNil(t, input)

		require.Len(t, input.SermonNotes, 1)
		assert.Equal(t, "note-1", input.SermonNotes[0].ID)
		assert.Equal(t, "Test Note", input.SermonNotes[0].Title)
		assert.Equal(t, "hash-abc", input.SermonNotes[0].ContentHash)

		require.Len(t, input.Gatherings, 1)
		assert.Equal(t, "g-1", input.Gatherings[0].ID)
		assert.Equal(t, "Sunday Morning", input.Gatherings[0].Name)
		assert.Equal(t, "#FFFFFF", input.Gatherings[0].ColorHex)

		require.Len(t, input.SermonSeries, 1)
		assert.Equal(t, "s-1", input.SermonSeries[0].ID)
		assert.Equal(t, "Romans", input.SermonSeries[0].Name)

		require.Len(t, input.Preachers, 1)
		assert.Equal(t, "p-1", input.Preachers[0].ID)
		assert.Equal(t, "Pastor John", input.Preachers[0].Name)

		require.Len(t, input.Tags, 1)
		assert.Equal(t, "t-1", input.Tags[0].ID)
		assert.Equal(t, "Theology", input.Tags[0].Name)
	})
}

func TestPushSyncChangesResultToResponse(t *testing.T) {
	t.Run(
		"nil result returns empty response with non-nil slices",
		func(t *testing.T) {
			res := converter.PushSyncChangesResultToResponse(nil)
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
		result := &pushsyncchanges.PushSyncChangesResult{
			SyncedAt:          syncedAt,
			MissingContentIDs: []string{"n1", "n2"},
			Conflicts:         []string{"n3"},
		}

		res := converter.PushSyncChangesResultToResponse(result)
		assert.NotNil(t, res)
		assert.Equal(t, syncedAt, res.SyncedAt)
		assert.Equal(t, []string{"n1", "n2"}, res.MissingContentIDs)
		assert.Equal(t, []string{"n3"}, res.Conflicts)
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
			result := &pullsyncchanges.SyncPullChangesResult{
				ServerTime: now,
				SermonNotes: []*entity.SermonNote{
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
				SermonSeries: []*entity.SermonSeries{
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
