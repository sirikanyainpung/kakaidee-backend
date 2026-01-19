package productStock

import (
	"kakaidee-backend/internal/helper"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (c *productStockController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	now := time.Now()

	products, err := c.svc.Get(ctx.Request().Context(), keyword, now)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get lot failed."),
		)
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Get lot success.", products),
	)
}
