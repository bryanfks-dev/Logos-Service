package http

import (
	nethttp "net/http"
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
// @Tags         API / Sync
// @Accept       json
// @Produce      json
// @Param        request body model.SyncPushRequest true "Sync push batch request payload"
// @Success      200 {object} model.APIResponse[model.SyncPushResponse]
// @Failure      422 {object} model.APIResponse[any]
// @Failure      409 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/sync/push [post]
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

	if err := req.Validate(); err != nil {
		h.log.Warn(
			"failed to validate sync push request",
			zap.Error(err),
			zap.String("request_id", requestID),
		)

		return err
	}

	batch := converter.SyncPushRequestToBatch(&req)

	result, err := h.syncUsecase.Push(ctx.Request().Context(), batch)
	if err != nil {
		return err
	}

	responseData := converter.SyncPushResultToResponse(result)
	response := builder.NewAPIResponseBuilder[model.SyncPushResponse]().
		SetRequestID(requestID).
		SetData(responseData).
		Build()

	return ctx.JSON(nethttp.StatusOK, response)
}

// Pull godoc
// @Summary      Sync pull delta
// @Description  Pull records updated since specified datetime across all entities
// @Tags         API / Sync
// @Accept       json
// @Produce      json
// @Param        since query string false "RFC3339 or YYYY-MM-DD datetime to query records updated since"
// @Success      200 {object} model.APIResponse[model.SyncPullResponse]
// @Failure      422 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/sync/pull [get]
func (h *SyncHandler) Pull(ctx *echo.Context) error {
	requestID := correlation.RequestID(ctx)

	var since time.Time
	if sinceStr := ctx.QueryParam("since"); sinceStr != "" {
		parsed, err := parseSinceParam(sinceStr)
		if err != nil {
			h.log.Warn(
				"failed to parse since query parameter",
				zap.Error(err),
				zap.String("since", sinceStr),
				zap.String("request_id", requestID),
			)

			return err
		}

		since = parsed
	}

	result, err := h.syncUsecase.Pull(ctx.Request().Context(), since)
	if err != nil {
		return err
	}

	responseData := converter.SyncPullResultToResponse(result)
	response := builder.NewAPIResponseBuilder[model.SyncPullResponse]().
		SetRequestID(requestID).
		SetData(responseData).
		Build()

	return ctx.JSON(nethttp.StatusOK, response)
}

func parseSinceParam(sinceStr string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, sinceStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, apperror.ErrInvalidPayload
}
