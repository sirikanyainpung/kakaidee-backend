package category

import (
	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/usecase/category"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type CategoryController struct {
	svc *category.CategoryService
}

func NewCategoryController(svc *category.CategoryService) *CategoryController {
	return &CategoryController{svc: svc}
}

func (c *CategoryController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	now := time.Now()

	categorys, err := c.svc.Get(ctx.Request().Context(), keyword, now)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get category failed."),
		)
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Get category success.", categorys),
	)
}
