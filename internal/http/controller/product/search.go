package product

import (
	"kakaidee-backend/internal/helper"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// func (c *ProductController) SearchV1(ctx echo.Context) error {

// 	keyword := ctx.QueryParam("keyword")
// 	loc, _ := time.LoadLocation("Asia/Bangkok")
// 	now := time.Now().In(loc)

// 	products, err := c.svc.Get(ctx.Request().Context(), keyword, now)
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError,
// 			helper.Error(500, "error", "Get product failed."),
// 		)
// 	}

// 	return ctx.JSON(
// 		http.StatusOK,
// 		helper.Success("Get product success.", products),
// 	)
// }

func (c *ProductController) GetV2(ctx echo.Context) error {
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

	pageStr := ctx.QueryParam("page")
	limitStr := ctx.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	helper.PrintStructJson(" ----- page ----- ")
	helper.PrintStructJson(page)
	helper.PrintStructJson(" ----- limit ----- ")
	helper.PrintStructJson(limit)
	helper.PrintStructJson(" ----- skip ----- ")
	helper.PrintStructJson(skip)

	// return nil

	result, err := c.svc.Get(
		ctx.Request().Context(),
		req.Keyword,
		req.SKUCode,
		req.CategoryCode,
		req.WarehouseName,
		req.LotNo,
		req.Status,
		now,
		page,
		limit,
		skip,
	)
	if err != nil {
		helper.PrintStructJson(" ----- err ----- ")
		helper.PrintStructJson(err)
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get product failed."),
		)
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Get product success.", result),
	)
}
