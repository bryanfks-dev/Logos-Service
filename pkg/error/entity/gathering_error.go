package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrGatheringCreationFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"gathering_creation_failed",
		"failed to create gathering",
		cause,
	)
}

func ErrGatheringUpdateFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"gathering_update_failed",
		"failed to update gathering",
		cause,
	)
}

func ErrGatheringRetrievalFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"gathering_retrieval_failed",
		"failed to retrieve gathering",
		cause,
	)
}

func ErrGatheringDeleteFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"gathering_delete_failed",
		"failed to delete gathering",
		cause,
	)
}

func ErrGatheringNotFound(id string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"gathering_not_found",
		"gathering not found: "+id,
	)
}
