package product

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/usecase/product"
)

func (c *ProductController) ImportExcel(ctx echo.Context) error {
	ctx.Logger().Info("👉 Product Import Excel called")

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest,
			helper.Error(400, "bad request", "Excel file is required."),
		)
	}

	result, err := c.svc.ImportFromExcel(
		ctx.Request().Context(),
		fileHeader,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", err.Error()),
		)
	}

	return ctx.JSON(
		http.StatusCreated,
		helper.Create("Import product success.", result),
	)
}