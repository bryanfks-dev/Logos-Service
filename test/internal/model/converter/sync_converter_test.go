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
)

func TestSyncPushChangesRequestToSyncChangesBatch(t *testing.T) {
	t.Run("nil request returns empty batch", func(t *testing.T) {
		batch := converter.SyncPushChangesRequestToSyncChangesBatch(nil)
		assert.NotNil(t, batch)
		assert.Empty(t, batch.SermonNotes)
		assert.Empty(t, batch.Gatherings)
		assert.Empty(t, batch.SermonSeries)
		assert.Empty(t, batch.Preachers)
		assert.Empty(t, batch.Tags)
	})

	t.Run("converts full sync push request to batch", func(t *testing.T) {
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

		batch := converter.SyncPushChangesRequestToSyncChangesBatch(req)
		assert.NotNil(t, batch)

		require.Len(t, batch.SermonNotes, 1)
		assert.Equal(t, "note-1", batch.SermonNotes[0].ID)
		assert.Equal(t, "Test Note", batch.SermonNotes[0].Title)
		assert.Equal(t, "hash-abc", batch.SermonNotes[0].ContentHash)

		require.Len(t, batch.Gatherings, 1)
		assert.Equal(t, "g-1", batch.Gatherings[0].ID)
		assert.Equal(t, "Sunday Morning", batch.Gatherings[0].Name)
		assert.Equal(t, "#FFFFFF", batch.Gatherings[0].ColorHex)

		require.Len(t, batch.SermonSeries, 1)
		assert.Equal(t, "s-1", batch.SermonSeries[0].ID)
		assert.Equal(t, "Romans", batch.SermonSeries[0].Name)

		require.Len(t, batch.Preachers, 1)
		assert.Equal(t, "p-1", batch.Preachers[0].ID)
		assert.Equal(t, "Pastor John", batch.Preachers[0].Name)

		require.Len(t, batch.Tags, 1)
		assert.Equal(t, "t-1", batch.Tags[0].ID)
		assert.Equal(t, "Theology", batch.Tags[0].Name)
	})
}

func TestSyncPushChangesResultToResponse(t *testing.T) {
	t.Run(
		"nil result returns empty response with non-nil slices",
		func(t *testing.T) {
			res := converter.SyncPushChangesResultToResponse(nil)
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

		res := converter.SyncPushChangesResultToResponse(result)
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
			result := &model.SyncPullChangesResult{
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
