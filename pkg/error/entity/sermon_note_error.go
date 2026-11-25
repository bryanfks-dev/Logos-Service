package entityerror

import (
	apperror "github.com/bryanfks-dev/logos/pkg/error"
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

func ErrSermonNoteRetrievalFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_retrieval_failed",
		"failed to retrieve sermon note",
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
