package models

import "time"

type Product struct {
	Barcode            string    `json:"barcode" bson:"barcode"`
	SKUCode            string    `json:"sku_code" bson:"sku_code"`
	ProductName        string    `json:"product_name" bson:"product_name"`
	ProductDescription string    `json:"product_description" bson:"product_description"`
	CategoryCode       int32     `json:"category_code" bson:"category_code"`
	SupplierCode       int32     `json:"supplier_code" bson:"supplier_code"`
	BrandCode          int32     `json:"brand_code" bson:"brand_code"`
	BalanceQty         int       `json:"balance_qty" bson:"balance_qty"`
	Unit               string    `json:"unit" bson:"unit"`
	CostPrice          float64   `json:"cost_price" bson:"cost_price"`
	Status             string    `json:"status" bson:"status"`
	CreatedBy          string    `json:"created_by" bson:"created_by"`
	UpdatedBy          string    `json:"updated_by" bson:"updated_by"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
}

func (Product) CollectionName() string {
	return "product_masters"
}
