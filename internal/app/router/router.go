package router

import (
	ProductController "kakaidee-backend/internal/http/controller/product"
	StaffsController "kakaidee-backend/internal/http/controller/staffs"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	Staffs  *StaffsController.StaffsController
	Product *ProductController.ProductController
}

func Register(e *echo.Echo, h Handlers) {

	// ===== Staffs =====
	staffs := e.Group("/staffs")
	staffs.POST("", h.Staffs.Create)

	// ===== Products =====
	product := e.Group("/product")
	product.POST("/create", h.Product.Create)
	product.GET("", h.Product.Search)
}
