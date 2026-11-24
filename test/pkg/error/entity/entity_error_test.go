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
			entityerror.ErrGatheringFindByIDFailed,
			"gathering_find_by_id_failed",
		},
		{
			entityerror.ErrGatheringFindAllFailed,
			"gathering_find_all_failed",
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
			entityerror.ErrSermonSeriesFindByIDFailed,
			"sermon_series_find_by_id_failed",
		},
		{
			entityerror.ErrSermonSeriesFindAllFailed,
			"sermon_series_find_all_failed",
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
			entityerror.ErrPreacherFindByIDFailed,
			"preacher_find_by_id_failed",
		},
		{
			entityerror.ErrPreacherFindAllFailed,
			"preacher_find_all_failed",
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
		{entityerror.ErrTagFindAllFailed, "tag_find_all_failed"},
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
			entityerror.ErrSermonNoteFindByIDFailed,
			"sermon_note_find_by_id_failed",
		},
		{
			entityerror.ErrSermonNoteFindAllFailed,
			"sermon_note_find_all_failed",
		},
		{
			entityerror.ErrSermonNoteSearchFailed,
			"sermon_note_search_failed",
		},
		{
			entityerror.ErrSermonNoteFindByScriptureFailed,
			"sermon_note_find_by_scripture_failed",
		},
		{
			entityerror.ErrSermonNoteDeleteFailed,
			"sermon_note_delete_failed",
		},

		// Bible errors
		{
			entityerror.ErrBibleGetVersesFailed,
			"bible_get_verses_failed",
		},
		{
			entityerror.ErrBibleSearchByTextFailed,
			"bible_search_by_text_failed",
		},
		{
			entityerror.ErrBibleImportTranslationPackFailed,
			"bible_import_translation_pack_failed",
		},
		{
			entityerror.ErrBibleListAvailableTranslationsFailed,
			"bible_list_available_translations_failed",
		},

		// SyncOutbox errors
		{
			entityerror.ErrSyncOutboxEnqueueFailed,
			"sync_outbox_enqueue_failed",
		},
		{
			entityerror.ErrSyncOutboxGetPendingBatchFailed,
			"sync_outbox_get_pending_batch_failed",
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
