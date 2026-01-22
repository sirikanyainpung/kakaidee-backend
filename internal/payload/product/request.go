package product

import "time"

type CreateProductRequest struct {
	Barcode            string    `json:"barcode" bson:"barcode"`
	SKUCode            string    `json:"sku_code" bson:"sku_code"`
	ProductName        string    `json:"product_name" bson:"product_name"`
	ProductDescription string    `json:"product_description" bson:"product_description"`
	CategoryCode       int32     `json:"category_code" bson:"category_code"`
	SupplierCode       int32     `json:"supplier_code" bson:"supplier_code"`
	BrandCode          int32     `json:"brand_code" bson:"brand_code"`
	Unit               int32     `json:"unit" bson:"unit"`
	CostPrice          float64   `json:"cost_price" bson:"cost_price"`
	LotNo              string    `json:"lot_no" bson:"lots_no"`
	WarehouseName      string    `json:"warehouse_name" bson:"warehouses_name"`
	WarehouseZone      string    `json:"warehouse_zone" bson:"warehouses_zone"`
	Bin                string    `json:"bin" bson:"bin"`
	StockType          string    `json:"stock_type" bson:"stock_type"`
	ReceiveQty         int32     `json:"receive_qty" bson:"receive_qty"`
	MFG                time.Time `json:"mfg" bson:"mfg"`
	EXP                time.Time `json:"exp" bson:"exp"`
	CreatedBy          string    `json:"created_by" bson:"created_by"`
	UpdatedBy          string    `json:"updated_by" bson:"updated_by"`
}

type ImportProductRequest struct {
	Barcode            string    `json:"barcode" bson:"barcode"`
	SKUCode            string    `json:"sku_code" bson:"sku_code"`
	ProductName        string    `json:"product_name" bson:"product_name"`
	ProductDescription string    `json:"product_description" bson:"product_description"`
	CategoryCode       int32     `json:"category_code" bson:"category_code"`
	SupplierCode       int32     `json:"supplier_code" bson:"supplier_code"`
	BrandCode          int32     `json:"brand_code" bson:"brand_code"`
	Unit               int32     `json:"unit" bson:"unit"`
	CostPrice          float64   `json:"cost_price" bson:"cost_price"`
	LotNo              string    `json:"lot_no" bson:"lots_no"`
	WarehouseName      string    `json:"warehouse_name" bson:"warehouses_name"`
	WarehouseZone      string    `json:"warehouse_zone" bson:"warehouses_zone"`
	Bin                string    `json:"bin" bson:"bin"`
	StockType          string    `json:"stock_type" bson:"stock_type"`
	ReceiveQty         int32     `json:"receive_qty" bson:"receive_qty"`
	MFG                time.Time `json:"mfg" bson:"mfg"`
	EXP                time.Time `json:"exp" bson:"exp"`
}
