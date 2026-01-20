package product

import (
	"kakaidee-backend/internal/helper"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (c *ProductController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

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

func (c *ProductController) Get(ctx echo.Context) error {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	req := struct {
		Keyword       string `query:"keyword"`
		SKUCode       string `query:"sku_code"`
		CategoryCode  string `query:"category_code"`
		WarehouseName string `query:"warehouse_name"`
		LotNo         string `query:"lot_no"`
		Status        string `query:"status"`
	}{}

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest,
			helper.Error(400, "bad request", "invalid query param"),
		)
	}

	result, err := c.svc.GetV2(
		ctx.Request().Context(),
		req.Keyword,
		req.SKUCode,
		req.CategoryCode,
		req.WarehouseName,
		req.LotNo,
		req.Status,
		now,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get product failed."),
		)
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Get product success.", result),
	)
}
