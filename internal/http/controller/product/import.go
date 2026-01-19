package product

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"kakaidee-backend/internal/helper"
)

func (c *ProductController) ImportExcel(ctx echo.Context) error {
	ctx.Logger().Info("👉 Product Import Excel called")

	// ===== get file =====
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest,
			helper.Error(400, "bad request", "Excel file is required."),
		)
	}

	// ===== mock created by (ปรับตาม auth จริงได้) =====
	createdBy := "admin"
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	// now := time.Now()

	result, err := c.svc.ImportFromExcel(
		ctx.Request().Context(),
		fileHeader,
		createdBy,
		now,
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
