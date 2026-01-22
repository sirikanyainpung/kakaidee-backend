package router

import (
	BrandController "kakaidee-backend/internal/http/controller/brand"
	CategoryController "kakaidee-backend/internal/http/controller/category"
	ProductController "kakaidee-backend/internal/http/controller/product"
	ProductStockController "kakaidee-backend/internal/http/controller/product_stock"
	StaffsController "kakaidee-backend/internal/http/controller/staffs"
	SupplierController "kakaidee-backend/internal/http/controller/supplier"
	TransactionLogController "kakaidee-backend/internal/http/controller/transaction_log"
	UnitController "kakaidee-backend/internal/http/controller/unit"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	Staffs         *StaffsController.StaffsController
	Product        *ProductController.ProductController
	ProductStock   *ProductStockController.ProductStockController
	Brand          *BrandController.BrandController
	Category       *CategoryController.CategoryController
	Supplier       *SupplierController.SupplierController
	Unit           *UnitController.UnitController
	TransactionLog *TransactionLogController.TransactionLogController
}

func Register(e *echo.Echo, h Handlers) {

	// ===== Staffs =====
	staffs := e.Group("/staffs")
	staffs.POST("", h.Staffs.Create)

	// ===== Products =====
	product := e.Group("/product")
	product.GET("", h.Product.Get)
	product.POST("/create", h.Product.Create)
	product.GET("/export", h.Product.ExportV2)

	product.POST("/import/excel", h.Product.ImportExcelV2)  // ImportExcel
	product.POST("/import/excel-v2", h.Product.ImportExcel) // ImportExcelV2

	product.GET("/all-product", h.Product.GetV2)
	product.GET("/export/all-product", h.Product.ExportV2)

	// ===== Products Stock =====
	productStock := e.Group("/product-stocks")
	productStock.GET("/search/lot", h.ProductStock.GetByLotNo)
	productStock.GET("/search/warehouse", h.ProductStock.GetByWarehouseName)

	// ===== Brand =====
	brand := e.Group("/brand")
	brand.GET("", h.Brand.Search)

	// ===== Category =====
	category := e.Group("/category")
	category.GET("", h.Category.Search)

	// ===== Supplier =====
	supplier := e.Group("/supplier")
	supplier.GET("", h.Supplier.Search)

	// ===== Unit =====
	unit := e.Group("/unit")
	unit.GET("", h.Unit.Search)

	// ===== Transaction Log =====
	transactionLog := e.Group("/transaction-log")
	transactionLog.GET("", h.TransactionLog.GetTransactionLog)
	transactionLog.GET("/export", h.TransactionLog.ExportTransactionLog)

}
