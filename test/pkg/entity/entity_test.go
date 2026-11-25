package entity_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/selah/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertUUIDv7(t *testing.T, id string) {
	t.Helper()
	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, byte(7), parsed[6]>>4, "expected UUID version 7")
}

func TestTableNames(t *testing.T) {
	assert.Equal(t, "gatherings", entity.Gathering{}.TableName())
	assert.Equal(t, "sermon_series", entity.SermonSeries{}.TableName())
	assert.Equal(t, "preachers", entity.Preacher{}.TableName())
	assert.Equal(t, "tags", entity.Tag{}.TableName())
	assert.Equal(t, "sermon_notes", entity.SermonNote{}.TableName())
	assert.Equal(t, "note_verses", entity.NoteVerse{}.TableName())
	assert.Equal(t, "note_tags", entity.NoteTag{}.TableName())
	assert.Equal(t, "notes_fts", entity.NotesFTS{}.TableName())
	assert.Equal(t, "bible_verses", entity.BibleVerse{}.TableName())
	assert.Equal(t, "bible_fts", entity.BibleFTS{}.TableName())
	assert.Equal(t, "sync_outboxes", entity.SyncOutbox{}.TableName())
}

func TestBeforeCreateHooks(t *testing.T) {
	t.Run("Gathering BeforeCreate generates UUID if empty", func(t *testing.T) {
		g := &entity.Gathering{}
		assert.NoError(t, g.BeforeCreate(nil))
		assert.NotEmpty(t, g.ID)
		assertUUIDv7(t, g.ID)

		gExisting := &entity.Gathering{ID: "custom-id"}
		assert.NoError(t, gExisting.BeforeCreate(nil))
		assert.Equal(t, "custom-id", gExisting.ID)
	})

	t.Run(
		"SermonSeries BeforeCreate generates UUID if empty",
		func(t *testing.T) {
			s := &entity.SermonSeries{}
			assert.NoError(t, s.BeforeCreate(nil))
			assert.NotEmpty(t, s.ID)
			assertUUIDv7(t, s.ID)

			sExisting := &entity.SermonSeries{ID: "series-id"}
			assert.NoError(t, sExisting.BeforeCreate(nil))
			assert.Equal(t, "series-id", sExisting.ID)
		},
	)

	t.Run("Preacher BeforeCreate generates UUID if empty", func(t *testing.T) {
		p := &entity.Preacher{}
		assert.NoError(t, p.BeforeCreate(nil))
		assert.NotEmpty(t, p.ID)
		assertUUIDv7(t, p.ID)

		pExisting := &entity.Preacher{ID: "preacher-id"}
		assert.NoError(t, pExisting.BeforeCreate(nil))
		assert.Equal(t, "preacher-id", pExisting.ID)
	})

	t.Run("Tag BeforeCreate generates UUID if empty", func(t *testing.T) {
		tag := &entity.Tag{}
		assert.NoError(t, tag.BeforeCreate(nil))
		assert.NotEmpty(t, tag.ID)
		assertUUIDv7(t, tag.ID)

		tagExisting := &entity.Tag{ID: "tag-id"}
		assert.NoError(t, tagExisting.BeforeCreate(nil))
		assert.Equal(t, "tag-id", tagExisting.ID)
	})

	t.Run(
		"SermonNote BeforeCreate generates UUID and FilePath if empty",
		func(t *testing.T) {
			note := &entity.SermonNote{}
			assert.NoError(t, note.BeforeCreate(nil))
			assert.NotEmpty(t, note.ID)
			assertUUIDv7(t, note.ID)
			assert.NotEmpty(t, note.FilePath)
			assert.Contains(t, note.FilePath, note.ID+".json")

			noteExisting := &entity.SermonNote{
				ID:       "note-123",
				FilePath: "/custom/path/note.json",
			}
			assert.NoError(t, noteExisting.BeforeCreate(nil))
			assert.Equal(t, "note-123", noteExisting.ID)
			assert.Equal(t, "/custom/path/note.json", noteExisting.FilePath)
		},
	)

	t.Run("NoteVerse BeforeCreate generates UUID if empty", func(t *testing.T) {
		nv := &entity.NoteVerse{}
		assert.NoError(t, nv.BeforeCreate(nil))
		assert.NotEmpty(t, nv.ID)
		assertUUIDv7(t, nv.ID)

		nvExisting := &entity.NoteVerse{ID: "nv-id"}
		assert.NoError(t, nvExisting.BeforeCreate(nil))
		assert.Equal(t, "nv-id", nvExisting.ID)
	})

	t.Run(
		"BibleVerse BeforeCreate generates UUID if empty",
		func(t *testing.T) {
			bv := &entity.BibleVerse{}
			assert.NoError(t, bv.BeforeCreate(nil))
			assert.NotEmpty(t, bv.ID)
			assertUUIDv7(t, bv.ID)

			bvExisting := &entity.BibleVerse{ID: "bv-id"}
			assert.NoError(t, bvExisting.BeforeCreate(nil))
			assert.Equal(t, "bv-id", bvExisting.ID)
		},
	)

	t.Run(
		"SyncOutbox BeforeCreate generates UUID and sets QueuedAt",
		func(t *testing.T) {
			outbox := &entity.SyncOutbox{}
			assert.NoError(t, outbox.BeforeCreate(nil))
			assert.NotEmpty(t, outbox.ID)
			assertUUIDv7(t, outbox.ID)
			assert.False(t, outbox.QueuedAt.IsZero())

			fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			outboxExisting := &entity.SyncOutbox{
				ID:       "outbox-id",
				QueuedAt: fixedTime,
			}
			assert.NoError(t, outboxExisting.BeforeCreate(nil))
			assert.Equal(t, "outbox-id", outboxExisting.ID)
			assert.Equal(t, fixedTime, outboxExisting.QueuedAt)
		},
	)
}

func TestSermonNote_ComparisonMethods(t *testing.T) {
	t.Run(
		"HasConflictWith handles nil and version/hash checks",
		func(t *testing.T) {
			var nilNote *entity.SermonNote
			incoming := &entity.SermonNote{Version: 2, ContentHash: "hash-a"}
			assert.False(t, nilNote.HasConflictWith(incoming))

			existing := &entity.SermonNote{Version: 3, ContentHash: "hash-b"}
			assert.False(t, existing.HasConflictWith(nil))
			assert.True(t, existing.HasConflictWith(incoming))

			// Same content hash -> no conflict even if existing version is higher
			assert.False(
				t,
				existing.HasConflictWith(&entity.SermonNote{
					Version:     2,
					ContentHash: "hash-b",
				}),
			)

			// Incoming version >= existing version -> no conflict
			assert.False(
				t,
				existing.HasConflictWith(&entity.SermonNote{
					Version:     3,
					ContentHash: "hash-a",
				}),
			)
		},
	)

	t.Run(
		"IsContentMissingFor compares ContentHash and hydrated ContentJSON",
		func(t *testing.T) {
			var nilNote *entity.SermonNote
			incoming := &entity.SermonNote{ContentHash: "hash-a"}
			assert.False(t, nilNote.IsContentMissingFor(incoming))

			existing := &entity.SermonNote{ContentHash: "hash-a"}
			assert.False(t, existing.IsContentMissingFor(nil))
			assert.False(t, existing.IsContentMissingFor(incoming))
			assert.True(
				t,
				existing.IsContentMissingFor(&entity.SermonNote{
					ContentHash: "hash-other",
				}),
			)

			content := `{"ops":[{"insert":"hello"}]}`
			sum := sha256.Sum256([]byte(content))
			contentHash := hex.EncodeToString(sum[:])

			hydrated := &entity.SermonNote{
				ContentHash: "stale-db-hash",
				ContentJSON: content,
			}
			assert.False(
				t,
				hydrated.IsContentMissingFor(&entity.SermonNote{
					ContentHash: contentHash,
				}),
			)
			assert.True(
				t,
				hydrated.IsContentMissingFor(&entity.SermonNote{
					ContentHash: "stale-db-hash",
				}),
			)
		},
	)
}
