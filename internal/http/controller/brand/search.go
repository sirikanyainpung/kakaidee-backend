package brand

import (
	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/usecase/brand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type BrandController struct {
	svc *brand.BrandService
}

func NewBrandController(svc *brand.BrandService) *BrandController {
	return &BrandController{svc: svc}
}

func (c *BrandController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	now := time.Now()

	brands, err := c.svc.Get(ctx.Request().Context(), keyword, now)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get brand failed."),
		)
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Get brand success.", brands),
	)
}
