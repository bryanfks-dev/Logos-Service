package entityerror_test

import (
	"errors"
	"testing"

	apperror "github.com/bryanfks-dev/selah/pkg/error"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"github.com/stretchr/testify/assert"
)

func TestEntityErrors(t *testing.T) {
	cause := errors.New("underlying cause error")

	errs := []struct {
		errFn func(error) *apperror.AppError
		code  string
	}{
		// Gathering errors
		{
			entityerror.ErrGatheringCreationFailed,
			"gathering_creation_failed",
		},
		{
			entityerror.ErrGatheringUpdateFailed,
			"gathering_update_failed",
		},
		{
			entityerror.ErrGatheringRetrievalFailed,
			"gathering_retrieval_failed",
		},
		{
			entityerror.ErrGatheringDeleteFailed,
			"gathering_delete_failed",
		},

		// SermonSeries errors
		{
			entityerror.ErrSermonSeriesCreationFailed,
			"sermon_series_creation_failed",
		},
		{
			entityerror.ErrSermonSeriesUpdateFailed,
			"sermon_series_update_failed",
		},
		{
			entityerror.ErrSermonSeriesRetrievalFailed,
			"sermon_series_retrieval_failed",
		},
		{
			entityerror.ErrSermonSeriesDeleteFailed,
			"sermon_series_delete_failed",
		},

		// Preacher errors
		{
			entityerror.ErrPreacherCreationFailed,
			"preacher_creation_failed",
		},
		{
			entityerror.ErrPreacherUpdateFailed,
			"preacher_update_failed",
		},
		{
			entityerror.ErrPreacherRetrievalFailed,
			"preacher_retrieval_failed",
		},
		{
			entityerror.ErrPreacherDeleteFailed,
			"preacher_delete_failed",
		},

		// Tag errors
		{
			entityerror.ErrTagCreationFailed,
			"tag_creation_failed",
		},
		{entityerror.ErrTagUpdateFailed, "tag_update_failed"},
		{entityerror.ErrTagRetrievalFailed, "tag_retrieval_failed"},
		{entityerror.ErrTagDeleteFailed, "tag_delete_failed"},

		// SermonNote errors
		{
			entityerror.ErrSermonNoteCreationFailed,
			"sermon_note_creation_failed",
		},
		{
			entityerror.ErrSermonNoteUpdateFailed,
			"sermon_note_update_failed",
		},
		{
			entityerror.ErrSermonNoteRetrievalFailed,
			"sermon_note_retrieval_failed",
		},
		{
			entityerror.ErrSermonNoteDeleteFailed,
			"sermon_note_delete_failed",
		},

		// Bible errors
		{
			entityerror.ErrBibleRetrievalFailed,
			"bible_retrieval_failed",
		},
		{
			entityerror.ErrBibleImportTranslationPackFailed,
			"bible_import_translation_pack_failed",
		},

		// SyncOutbox errors
		{
			entityerror.ErrSyncOutboxEnqueueFailed,
			"sync_outbox_enqueue_failed",
		},
		{
			entityerror.ErrSyncOutboxRetrievalFailed,
			"sync_outbox_retrieval_failed",
		},
		{
			entityerror.ErrSyncOutboxMarkSyncedFailed,
			"sync_outbox_mark_synced_failed",
		},
		{
			entityerror.ErrSyncOutboxClearSyncedFailed,
			"sync_outbox_clear_synced_failed",
		},
	}

	for _, tt := range errs {
		appErr := tt.errFn(cause)
		assert.Equal(t, tt.code, appErr.Code)
		assert.Equal(t, apperror.ErrKindInternal, appErr.Kind)
		assert.Equal(t, cause, appErr.Err)
		assert.Equal(t, cause, errors.Unwrap(appErr))
		assert.True(t, errors.Is(appErr, tt.errFn(nil)))
	}

	// Test NotFound errors
	assert.Equal(
		t,
		"gathering_not_found",
		entityerror.ErrGatheringNotFound("1").Code,
	)
	assert.Equal(
		t,
		apperror.ErrKindNotFound,
		entityerror.ErrGatheringNotFound("1").Kind,
	)

	assert.Equal(
		t,
		"sermon_series_not_found",
		entityerror.ErrSermonSeriesNotFound("1").Code,
	)
	assert.Equal(
		t,
		"preacher_not_found",
		entityerror.ErrPreacherNotFound("1").Code,
	)
	assert.Equal(
		t,
		"tag_not_found",
		entityerror.ErrTagNotFound("1").Code,
	)
	assert.Equal(
		t,
		"sermon_note_not_found",
		entityerror.ErrSermonNoteNotFound("1").Code,
	)
	assert.Equal(
		t,
		"bible_verse_not_found",
		entityerror.ErrBibleVerseNotFound("Gen 1:1").Code,
	)
}
