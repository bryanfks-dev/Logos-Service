package delivery

import (
	"net/http"

	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	_ "github.com/bryanfks-dev/selah/internal/model"
	modelbuilder "github.com/bryanfks-dev/selah/internal/model/builder"
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
	requestID := middlewarecorrelation.GetRequestID(ctx)

	response := modelbuilder.NewApiResponseBuilder[any]().
		SetRequestId(requestID).
		Build()

	return ctx.JSON(http.StatusOK, response)
}
