package delivery

import (
	"net/http"

	"github.com/bryanfks-dev/internal/delivery/middleware"
	_ "github.com/bryanfks-dev/internal/model"
	"github.com/bryanfks-dev/internal/model/builder"
	"github.com/labstack/echo/v5"
)

type EtcController struct{}

func NewEtcController() *EtcController {
	return &EtcController{}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Check service health status
// @Tags         API / etc
// @Accept       json
// @Produce      json
// @Success      200 {object} model.ApiResponse[any]
// @Router       /api/healthz [get]
func (c *EtcController) HealthCheck(ctx *echo.Context) error {
	requestID := middleware.GetRequestID(ctx)

	response := builder.NewApiResponseBuilder[any]().
		SetRequestId(requestID).
		Build()
	return ctx.JSON(http.StatusOK, response)
}
