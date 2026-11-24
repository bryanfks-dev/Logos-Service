package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrSermonNoteCreationFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_creation_failed",
		"failed to create sermon note",
		cause,
	)
}

func ErrSermonNoteUpdateFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_update_failed",
		"failed to update sermon note",
		cause,
	)
}

func ErrSermonNoteFindByIDFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_find_by_id_failed",
		"failed to find sermon note by id",
		cause,
	)
}

func ErrSermonNoteFindAllFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_find_all_failed",
		"failed to find sermon notes",
		cause,
	)
}

func ErrSermonNoteSearchFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_search_failed",
		"failed to search sermon notes",
		cause,
	)
}

func ErrSermonNoteFindByScriptureFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_find_by_scripture_failed",
		"failed to find sermon notes by scripture",
		cause,
	)
}

func ErrSermonNoteDeleteFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_delete_failed",
		"failed to delete sermon note",
		cause,
	)
}

func ErrSermonNoteNotFound(id string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"sermon_note_not_found",
		"sermon note not found: "+id,
	)
}
