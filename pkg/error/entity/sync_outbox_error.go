package entityerror

import (
	apperror "github.com/bryanfks-dev/selah/pkg/error"
)

func ErrSyncOutboxEnqueueFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sync_outbox_enqueue_failed",
		"failed to enqueue sync outbox entry",
		cause,
	)
}

func ErrSyncOutboxGetPendingBatchFailed(cause error) *apperror.AppError {
	return apperror.NewErrInternal(
		"sync_outbox_get_pending_batch_failed",
		"failed to get pending sync outbox batch",
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
