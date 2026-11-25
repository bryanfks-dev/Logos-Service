package http

import (
	nethttp "net/http"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	_ "github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/builder"
	"github.com/bryanfks-dev/logos/internal/model/converter"
	params "github.com/bryanfks-dev/logos/internal/model/param"
	"github.com/bryanfks-dev/logos/internal/model/request"
	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/internal/usecase"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type SyncController struct {
	pullSyncChangesUsecase *usecase.PullSyncChangesUsecase
	pushSyncChangesUsecase *usecase.PushSyncChangesUsecase
	log                    *zap.Logger
}

func NewSyncController(
	pullSyncChangesUsecase *usecase.PullSyncChangesUsecase,
	pushSyncChangesUsecase *usecase.PushSyncChangesUsecase,
	log *zap.Logger,
) *SyncController {
	return &SyncController{
		pullSyncChangesUsecase: pullSyncChangesUsecase,
		pushSyncChangesUsecase: pushSyncChangesUsecase,
		log:                    log,
	}
}

// Push godoc
// @Summary      Sync by pushing changes
// @Description  Push batch of entities to synchronize with server
// @Tags         API / Sync
// @Accept       json
// @Produce      json
// @Param        request body request.SyncPushChangesRequest true "Sync push batch request payload"
// @Success      200 {object} model.APIResponse[response.SyncPushChangesResponse]
// @Failure      400 {object} model.APIResponse[any]
// @Failure      422 {object} model.APIResponse[any]
// @Failure      409 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/sync/changes [post]
func (h *SyncController) Push(ctx *echo.Context) error {
	reqID := correlation.GetRequestID(ctx)

	req := new(request.SyncPushChangesRequest)
	if err := ctx.Bind(req); err != nil {
		h.log.Warn(
			"failed to bind sync push request",
			zap.Error(err),
			zap.String("request_id", reqID),
		)

		return apperror.NewErrRequestBodyMalformed()
	}

	if err := req.Validate(); err != nil {
		h.log.Warn(
			"failed to validate sync push request",
			zap.Error(err),
			zap.String("request_id", reqID),
		)

		return err
	}

	batch := converter.SyncPushChangesRequestToSyncChangesBatch(req)

	result, err := h.pushSyncChangesUsecase.Call(ctx.Request().Context(), batch)
	if err != nil {
		return err
	}

	resData := converter.SyncPushChangesResultToResponse(result)
	res := builder.NewAPIResponseBuilder[response.SyncPushChangesResponse]().
		SetRequestID(reqID).
		SetData(resData).
		Build()

	return ctx.JSON(nethttp.StatusOK, res)
}

// Pull godoc
// @Summary      Sync by pulling delta of changes
// @Description  Pull records updated since specified datetime across all entities
// @Tags         API / Sync
// @Accept       json
// @Produce      json
// @Param        since path string true "RFC3339 or YYYY-MM-DD datetime to query records updated since"
// @Success      200 {object} model.APIResponse[response.SyncPullChangesResponse]
// @Failure      400 {object} model.APIResponse[any]
// @Failure      422 {object} model.APIResponse[any]
// @Failure      500 {object} model.APIResponse[any]
// @Router       /api/sync/changes/{since} [get]
func (h *SyncController) Pull(ctx *echo.Context) error {
	reqID := correlation.GetRequestID(ctx)

	params := new(params.SyncPullChangesParam)
	if err := ctx.Bind(params); err != nil {
		h.log.Warn(
			"failed to bind sync pull params",
			zap.Error(err),
			zap.String("request_id", reqID),
		)

		return apperror.NewErrRequestParamsMalformed()
	}

	if err := params.Validate(); err != nil {
		h.log.Warn(
			"failed to validate sync pull params",
			zap.Error(err),
			zap.String("since", params.Since),
			zap.String("request_id", reqID),
		)

		return err
	}

	result, err := h.pullSyncChangesUsecase.Call(
		ctx.Request().Context(),
		params.ParsedSince(),
	)
	if err != nil {
		return err
	}

	resData := converter.SyncPullChangesResultToResponse(result)
	res := builder.NewAPIResponseBuilder[response.SyncPullChangesResponse]().
		SetRequestID(reqID).
		SetData(resData).
		Build()

	return ctx.JSON(nethttp.StatusOK, res)
}
