package entityerror

import (
	apperror "github.com/bryanfks-dev/logos/pkg/error"
)

func ErrSermonNoteAttachmentRetrievalFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_attachment_retrieval_failed",
		"failed to retrieve sermon note attachment",
		cause,
	)
}

func ErrSermonNoteAttachmentUpdateFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_note_attachment_update_failed",
		"failed to update sermon note attachment",
		cause,
	)
}

func ErrSermonNoteAttachmentNotFound(id string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"sermon_note_attachment_not_found",
		"sermon note attachment not found: "+id,
	)
}

func ErrSermonNoteAttachmentInvalidFile() *apperror.AppError {
	return apperror.NewErrBadRequest(
		"invalid_file",
		"attachment file invalid",
	)
}

func ErrSermonNoteAttachmentInvalidHash() *apperror.AppError {
	return apperror.NewErrBadRequest(
		"invalid_hash",
		"attachment content hash mismatch",
	)
}
