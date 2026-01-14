package product

import (
	"net/http"

	"github.com/labstack/echo/v4"

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
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	err := c.svc.Create(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, map[string]string{
		"status": "created",
	})
}

// func (c *ProductController) Create(ctx echo.Context) error {
// 	ctx.Logger().Info("👉 Product Create called")

// 	var req payloadProduct.CreateProductRequest
// 	if err := ctx.Bind(&req); err != nil {
// 		helper.PrintStructJson("--------------- err -----------------")
// 		helper.PrintStructJson(err)
// 		return ctx.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "invalid request body",
// 		})
// 	}

// 	helper.PrintStructJson("--------------- pass -----------------")
// 	helper.PrintStructJson(req)
// 	helper.PrintStructJson("--------------------------------")

// 	CreatedAt := time.Now()
// 	productStock := models.ProductStock{
// 		Barcode:       req.Barcode,
// 		SKUCode:       req.SKUCode,
// 		LotNo:         req.LotNo,
// 		WarehouseName: req.WarehouseName,
// 		WarehouseZone: req.WarehouseZone,
// 		Bin:           req.Bin,
// 		StockType:     req.StockType,
// 		ReceiveQty:    req.ReceiveQty,
// 		BalanceQty:    req.ReceiveQty,
// 		SellingQty:    0,
// 		MFG:           req.MFG,
// 		EXP:           req.EXP,
// 		Status:        "active",
// 		CreatedBy:     req.CreatedBy,
// 		UpdatedBy:     req.UpdatedBy,
// 		CreatedAt:     CreatedAt,
// 		UpdatedAt:     CreatedAt,
// 	}

// 	helper.PrintStructJson("---------- productStock ----------")
// 	helper.PrintStructJson(productStock)

// 	productMasters := models.Product{
// 		Barcode:            req.Barcode,
// 		SKUCode:            req.SKUCode,
// 		ProductName:        req.ProductName,
// 		ProductDescription: req.ProductDescription,
// 		CategoryCode:       req.CategoryCode,
// 		SupplierCode:       req.SupplierCode,
// 		BrandCode:          req.BrandCode,
// 		BalanceQty:         req.ReceiveQty,
// 		Unit:               req.Unit,
// 		CostPrice:          req.CostPrice,
// 		Status:             "active",
// 		CreatedBy:          req.CreatedBy,
// 		UpdatedBy:          req.UpdatedBy,
// 		CreatedAt:          CreatedAt,
// 		UpdatedAt:          CreatedAt,
// 	}

// 	helper.PrintStructJson("---------- productMasters ----------")
// 	helper.PrintStructJson(productMasters)

// 	ctxMongo := ctx.Request().Context()

// 	if c.db == nil {
// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{
// 			"error": "database not initialized",
// 		})
// 	}

// 	collectionProductStock := c.db.Collection(models.ProductStock{}.CollectionName())
// 	ctx.Logger().Info("👉 got collection ProductStock")
// 	collectionProduct := c.db.Collection(models.Product{}.CollectionName())
// 	ctx.Logger().Info("👉 got collection Product")

// 	// insert product_stock
// 	if _, err := collectionProductStock.
// 		InsertOne(ctxMongo, productStock); err != nil {

// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{
// 			"error": err.Error(),
// 		})
// 	}

// 	// insert product_stock
// 	if _, err := collectionProduct.
// 		InsertOne(ctxMongo, productMasters); err != nil {

// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{
// 			"error": err.Error(),
// 		})
// 	}

// 	return ctx.JSON(http.StatusCreated, map[string]string{
// 		"status": "created",
// 	})
// }
