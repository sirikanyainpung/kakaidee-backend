package supplier

import (
	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/usecase/supplier"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type SupplierController struct {
	svc *supplier.SupplierService
}

func NewSupplierController(svc *supplier.SupplierService) *SupplierController {
	return &SupplierController{svc: svc}
}

func (c *SupplierController) Search(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	now := time.Now()

	suppliers, err := c.svc.Get(ctx.Request().Context(), keyword, now)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Get supplier failed."),
		)
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Get supplier success.", suppliers),
	)
}
