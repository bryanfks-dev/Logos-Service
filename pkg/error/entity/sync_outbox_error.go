package entityerror

import (
	apperror "github.com/bryanfks-dev/logos/pkg/error"
)

func ErrSyncOutboxEnqueueFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sync_outbox_enqueue_failed",
		"failed to enqueue sync outbox entry",
		cause,
	)
}

func ErrSyncOutboxRetrievalFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sync_outbox_retrieval_failed",
		"failed to retrieve sync outbox entries",
		cause,
	)
}

func ErrSyncOutboxMarkSyncedFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sync_outbox_mark_synced_failed",
		"failed to mark sync outbox entries as synced",
		cause,
	)
}

func ErrSyncOutboxClearSyncedFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sync_outbox_clear_synced_failed",
		"failed to clear synced outbox entries",
		cause,
	)
}
