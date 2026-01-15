package product

import (
	"kakaidee-backend/internal/helper"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (c *ProductController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	now := time.Now()

	products, err := c.svc.Get(ctx.Request().Context(), keyword, now)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get product failed."),
		)
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Get product success.", products),
	)
}
