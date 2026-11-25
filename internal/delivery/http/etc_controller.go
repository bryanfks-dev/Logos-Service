package http

import (
	"net/http"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	_ "github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/builder"
	"github.com/labstack/echo/v5"
)

type EtcController struct{}

func NewEtcController() *EtcController {
	return &EtcController{}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Check service health status
// @Tags         API / Etc
// @Accept       json
// @Produce      json
// @Success      200 {object} model.APIResponse[any]
// @Router       /api/healthz [get]
func (c *EtcController) HealthCheck(ctx *echo.Context) error {
	reqID := correlation.GetRequestID(ctx)

	res := builder.NewAPIResponseBuilder[any]().
		SetRequestID(reqID).
		Build()

	return ctx.JSON(http.StatusOK, res)
}
