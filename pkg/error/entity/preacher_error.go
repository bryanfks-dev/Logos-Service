package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrPreacherCreationFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"preacher_creation_failed",
		"failed to create preacher",
		cause,
	)
}

func ErrPreacherUpdateFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"preacher_update_failed",
		"failed to update preacher",
		cause,
	)
}

func ErrPreacherRetrievalFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"preacher_retrieval_failed",
		"failed to retrieve preacher",
		cause,
	)
}

func ErrPreacherDeleteFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"preacher_delete_failed",
		"failed to delete preacher",
		cause,
	)
}

func ErrPreacherNotFound(id string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"preacher_not_found",
		"preacher not found: "+id,
	)
}
