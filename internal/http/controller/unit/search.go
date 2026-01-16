package unit

import (
	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/usecase/unit"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type UnitController struct {
	svc *unit.UnitService
}

func NewUnitController(svc *unit.UnitService) *UnitController {
	return &UnitController{svc: svc}
}

func (c *UnitController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	now := time.Now()

	units, err := c.svc.Get(ctx.Request().Context(), keyword, now)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get unit failed."),
		)
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Get unit success.", units),
	)
}
