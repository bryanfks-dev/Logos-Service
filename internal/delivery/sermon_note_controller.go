package delivery

import (
	"net/http"

	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/model"
	modelbuilder "github.com/bryanfks-dev/selah/internal/model/builder"
	"github.com/bryanfks-dev/selah/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type SermonNoteController struct {
	syncUsecase usecase.SyncUsecase
	log         *zap.Logger
}

func NewSermonNoteController(
	syncUsecase usecase.SyncUsecase,
	log *zap.Logger,
) *SermonNoteController {
	return &SermonNoteController{
		syncUsecase: syncUsecase,
		log:         log,
	}
}

// SyncNotes godoc
// @Summary      Sync sermon notes
// @Description  Synchronize notes between client and server with two-way synchronization
// @Tags         API / sermon-notes
// @Accept       json
// @Produce      json
// @Param        request body model.SyncNotesRequest true "Sync notes request payload"
// @Success      200 {object} model.ApiResponse[model.SyncNotesResponse]
// @Failure      400 {object} model.ApiResponse[any]
// @Failure      500 {object} model.ApiResponse[any]
// @Router       /api/sync/notes [post]
func (c *SermonNoteController) SyncNotes(ctx *echo.Context) error {
	requestID := middlewarecorrelation.GetRequestID(ctx)

	var req model.SyncNotesRequest
	if err := ctx.Bind(&req); err != nil {
		c.log.Warn(
			"failed to bind sync notes request",
			zap.Error(err),
			zap.String("request_id", requestID),
		)

		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	res, err := c.syncUsecase.SyncNotes(ctx.Request().Context(), &req)
	if err != nil {
		return err
	}

	response := modelbuilder.NewApiResponseBuilder[model.SyncNotesResponse]().
		SetRequestId(requestID).
		SetData(*res).
		Build()

	return ctx.JSON(http.StatusOK, response)
}
