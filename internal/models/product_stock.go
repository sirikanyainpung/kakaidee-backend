package models

import "time"

type ProductStock struct {
	Barcode       string    `json:"barcode" bson:"barcode"`
	SKUCode       string    `json:"sku_code" bson:"sku_code"`
	LotNo         string    `json:"lot_no" bson:"lot_no"`
	WarehouseName string    `json:"warehouse_name" bson:"warehouses_name"`
	WarehouseZone string    `json:"warehouse_zone" bson:"warehouses_zone"`
	Bin           string    `json:"bin" bson:"bin"`
	StockType     string    `json:"stock_type" bson:"stock_type"`
	ReceiveQty    int       `json:"receive_qty" bson:"receive_qty"`
	BalanceQty    int       `json:"balance_qty" bson:"balance_qty"`
	SellingQty    int       `json:"selling_qty" bson:"selling_qty"`
	MFG           time.Time `json:"mfg" bson:"mfg"`
	EXP           time.Time `json:"exp" bson:"exp"`
	Status        string    `json:"status" bson:"status"`
	CreatedBy     string    `json:"created_by" bson:"created_by"`
	UpdatedBy     string    `json:"updated_by" bson:"updated_by"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

func (ProductStock) CollectionName() string {
	return "product_stock"
}
