package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrTagCreationFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"tag_creation_failed",
		"failed to create tag",
		cause,
	)
}

func ErrTagUpdateFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"tag_update_failed",
		"failed to update tag",
		cause,
	)
}

func ErrTagFindAllFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"tag_find_all_failed",
		"failed to find all tags",
		cause,
	)
}

func ErrTagDeleteFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"tag_delete_failed",
		"failed to delete tag",
		cause,
	)
}

func ErrTagNotFound(id string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"tag_not_found",
		"tag not found: "+id,
	)
}
