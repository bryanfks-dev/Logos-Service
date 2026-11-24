package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrSermonSeriesCreationFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_series_creation_failed",
		"failed to create sermon series",
		cause,
	)
}

func ErrSermonSeriesUpdateFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_series_update_failed",
		"failed to update sermon series",
		cause,
	)
}

func ErrSermonSeriesFindByIDFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_series_find_by_id_failed",
		"failed to find sermon series by id",
		cause,
	)
}

func ErrSermonSeriesFindAllFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_series_find_all_failed",
		"failed to find all sermon series",
		cause,
	)
}

func ErrSermonSeriesDeleteFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sermon_series_delete_failed",
		"failed to delete sermon series",
		cause,
	)
}

func ErrSermonSeriesNotFound(id string) *apperror.AppError {
	return apperror.NewErrNotFound(
		"sermon_series_not_found",
		"sermon series not found: "+id,
	)
}
