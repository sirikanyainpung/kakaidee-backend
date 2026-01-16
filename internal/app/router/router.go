package router

import (
	BrandController "kakaidee-backend/internal/http/controller/brand"
	CategoryController "kakaidee-backend/internal/http/controller/category"
	ProductController "kakaidee-backend/internal/http/controller/product"
	StaffsController "kakaidee-backend/internal/http/controller/staffs"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	Staffs   *StaffsController.StaffsController
	Product  *ProductController.ProductController
	Brand    *BrandController.BrandController
	Category *CategoryController.CategoryController
}

func Register(e *echo.Echo, h Handlers) {

	// ===== Staffs =====
	staffs := e.Group("/staffs")
	staffs.POST("", h.Staffs.Create)

	// ===== Products =====
	product := e.Group("/product")
	product.POST("/create", h.Product.Create)
	product.GET("", h.Product.Search)

	// ===== Brand =====
	brand := e.Group("/brand")
	brand.GET("", h.Brand.Search)

	// ===== Category =====
	category := e.Group("/category")
	category.GET("", h.Category.Search)
}
