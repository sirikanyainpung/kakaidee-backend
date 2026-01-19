package product_stock

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/usecase/product_stock"
)

type ProductStockController struct {
	svc *product_stock.ProductStockService
}

func NewProductStockController(
	svc *product_stock.ProductStockService,
) *ProductStockController {
	return &ProductStockController{svc: svc}
}

// ===== GET by lot_no =====
func (c *ProductStockController) GetByLotNo(ctx echo.Context) error {
	lotNo := ctx.QueryParam("keyword")

	result, err := c.svc.GetLotNoList(
		ctx.Request().Context(),
		lotNo,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", err.Error()),
		)
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Get lot_no list success.", result),
	)
}

// ===== GET by warehouse_name =====
func (c *ProductStockController) GetByWarehouseName(ctx echo.Context) error {
	warehouse := ctx.QueryParam("keyword")
	if warehouse == "" {
		return ctx.JSON(http.StatusBadRequest,
			helper.Error(400, "bad request", "warehouse_name is required"),
		)
	}

	result, err := c.svc.GetByWarehouseName(
		ctx.Request().Context(),
		warehouse,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", err.Error()),
		)
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Get product stock by warehouse success.", result),
	)
}
