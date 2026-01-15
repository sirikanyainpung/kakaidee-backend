package product

import (
	"kakaidee-backend/internal/helper"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *ProductController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")

	products, err := c.svc.Get(ctx.Request().Context(), keyword)
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
