package product

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"kakaidee-backend/internal/helper"
	payloadProduct "kakaidee-backend/internal/payload/product"
	"kakaidee-backend/internal/usecase/product"
)

type ProductController struct {
	svc *product.ProductService
}

func NewProductsController(svc *product.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

func (c *ProductController) Create(ctx echo.Context) error {
	ctx.Logger().Info("👉 Product Create called")

	var req payloadProduct.CreateProductRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest,
			helper.Error(400, "bad request", "Invalid request body."),
		)
	}

	err := c.svc.Create(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Create product failed."),
		)
	}

	return ctx.JSON(
		http.StatusCreated,
		helper.Create("Create product success.", req),
	)
}
