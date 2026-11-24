package http

import (
	nethttp "net/http"
	"strconv"
	"time"

	"github.com/bryanfks-dev/selah/internal/delivery/http/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/model/builder"
	"github.com/bryanfks-dev/selah/internal/model/converter"
	"github.com/bryanfks-dev/selah/internal/usecase"
	apperror "github.com/bryanfks-dev/selah/pkg/error"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type SyncHandler struct {
	syncUsecase *usecase.SyncUsecase
	log         *zap.Logger
}

type SyncController = SyncHandler

func NewSyncHandler(
	syncUsecase *usecase.SyncUsecase,
	log *zap.Logger,
) *SyncHandler {
	return &SyncHandler{
		syncUsecase: syncUsecase,
		log:         log,
	}
}

func NewSyncController(
	syncUsecase *usecase.SyncUsecase,
	log *zap.Logger,
) *SyncController {
	return NewSyncHandler(syncUsecase, log)
}

// Push godoc
// @Summary      Sync push batch
// @Description  Push batch of entities to synchronize with server
// @Tags         API / sync
// @Accept       json
// @Produce      json
// @Param        request body model.SyncPushRequest true "Sync push batch request payload"
// @Success      200 {object} model.APIResponse[model.SyncPushResponse]
// @Failure      400 {object} model.APIResponse[any]
// @Failure      409 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/v1/sync/push [post]
func (h *SyncHandler) Push(ctx *echo.Context) error {
	requestID := correlation.RequestID(ctx)

	var req model.SyncPushRequest
	if err := ctx.Bind(&req); err != nil {
		h.log.Warn(
			"failed to bind sync push request",
			zap.Error(err),
			zap.String("request_id", requestID),
		)

		return apperror.ErrInvalidPayload
	}

	batch := converter.SyncPushRequestToBatch(&req)

	result, err := h.syncUsecase.Push(ctx.Request().Context(), batch)
	if err != nil {
		return err
	}

	resDTO := converter.SyncPushResultToResponse(result)

	var respData model.SyncPushResponse
	if resDTO != nil {
		respData = *resDTO
	}

	response := builder.NewAPIResponseBuilder[model.SyncPushResponse]().
		SetRequestID(requestID).
		SetData(respData).
		Build()

	return ctx.JSON(nethttp.StatusOK, response)
}

// Pull godoc
// @Summary      Sync pull delta
// @Description  Pull records updated since specified timestamp across all entities
// @Tags         API / sync
// @Accept       json
// @Produce      json
// @Param        since query int false "Timestamp (int64 or datetime) to query records updated since"
// @Success      200 {object} model.APIResponse[model.SyncPullResponse]
// @Failure      400 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/v1/sync/pull [get]
func (h *SyncHandler) Pull(ctx *echo.Context) error {
	requestID := correlation.RequestID(ctx)

	var since int64
	sinceStr := ctx.QueryParam("since")
	if sinceStr != "" {
		parsed, err := strconv.ParseInt(sinceStr, 10, 64)
		if err == nil {
			since = parsed
		} else {
			// Try parsing as ISO8601 / RFC3339 datetime
			t, err := time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				t, err = time.Parse("2006-01-02T15:04:05Z07:00", sinceStr)
			}
			if err != nil {
				t, err = time.Parse("2006-01-02", sinceStr)
			}
			if err != nil {
				h.log.Warn(
					"failed to parse since query parameter",
					zap.Error(err),
					zap.String("since", sinceStr),
					zap.String("request_id", requestID),
				)

				return apperror.ErrInvalidPayload
			}
			since = t.UnixMilli()
		}
	}

	result, err := h.syncUsecase.Pull(ctx.Request().Context(), since)
	if err != nil {
		return err
	}

	resDTO := converter.SyncPullResultToResponse(result)

	var respData model.SyncPullResponse
	if resDTO != nil {
		respData = *resDTO
	}

	response := builder.NewAPIResponseBuilder[model.SyncPullResponse]().
		SetRequestID(requestID).
		SetData(respData).
		Build()

	return ctx.JSON(nethttp.StatusOK, response)
}
