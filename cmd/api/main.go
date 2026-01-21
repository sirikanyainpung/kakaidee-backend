package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"kakaidee-backend/internal/app/router"
	"kakaidee-backend/internal/config"

	// controllers
	brandController "kakaidee-backend/internal/http/controller/brand"
	categoryController "kakaidee-backend/internal/http/controller/category"
	productController "kakaidee-backend/internal/http/controller/product"
	productStockController "kakaidee-backend/internal/http/controller/product_stock"
	staffsController "kakaidee-backend/internal/http/controller/staffs"
	supplierController "kakaidee-backend/internal/http/controller/supplier"
	transactionLogController "kakaidee-backend/internal/http/controller/transaction_log"
	unitController "kakaidee-backend/internal/http/controller/unit"

	// usecases
	brandUsecase "kakaidee-backend/internal/usecase/brand"
	categoryUsecase "kakaidee-backend/internal/usecase/category"
	productUsecase "kakaidee-backend/internal/usecase/product"
	productStockUsecase "kakaidee-backend/internal/usecase/product_stock"
	staffsUsecase "kakaidee-backend/internal/usecase/staffs"
	supplierUsecase "kakaidee-backend/internal/usecase/supplier"
	transactionLogUsecase "kakaidee-backend/internal/usecase/transaction_log"
	unitUsecase "kakaidee-backend/internal/usecase/unit"
)

func main() {
	// ===== Echo =====
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ===== Mongo Config =====
	mongoCfg := config.MongoConfig{
		URI:      os.Getenv("MONGODB_URI"),
		Database: os.Getenv("MONGODB_DB"),
	}

	db := config.NewMongoConnection(mongoCfg)

	// ===== CORS =====
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:4200",
		},
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
	}))

	// ===== Usecase =====
	productService := productUsecase.NewProductService(db)
	productStockService := productStockUsecase.NewProductStockService(db)
	staffsService := staffsUsecase.NewStaffsService()
	brandService := brandUsecase.NewBrandService(db)
	categoryService := categoryUsecase.NewCategoryService(db)
	supplierService := supplierUsecase.NewSupplierService(db)
	unitService := unitUsecase.NewUnitService(db)
	transactionLogService := transactionLogUsecase.NewTransactionLogService(db)

	// ===== Controller =====
	productCtrl := productController.NewProductsController(productService)
	productStockCtrl := productStockController.NewProductStockController(productStockService)
	staffsCtrl := staffsController.NewStaffsController(staffsService)
	brandCtrl := brandController.NewBrandController(brandService)
	categoryCtrl := categoryController.NewCategoryController(categoryService)
	supplierCtrl := supplierController.NewSupplierController(supplierService)
	unitCtrl := unitController.NewUnitController(unitService)
	ransactionLogCtrl := transactionLogController.NewTransactionLogController(transactionLogService)

	// ===== Router =====
	router.Register(e, router.Handlers{
		Product:        productCtrl,
		ProductStock:   productStockCtrl,
		Staffs:         staffsCtrl,
		Brand:          brandCtrl,
		Category:       categoryCtrl,
		Supplier:       supplierCtrl,
		Unit:           unitCtrl,
		TransactionLog: ransactionLogCtrl,
	})

	// ===== Start Server =====
	e.Logger.Fatal(e.Start(":8080"))
}
