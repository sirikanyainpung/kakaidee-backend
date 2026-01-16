package models

import "time"

type Supplier struct {
	SupplierCode int32     `json:"supplier_code" bson:"supplier_code"`
	SupplierName string    `json:"supplier_name" bson:"supplier_name"`
	Status       string    `json:"status" bson:"status"`
	Createbsony  string    `json:"created_by" bson:"created_by"`
	Updatebsony  string    `json:"updated_by" bson:"updated_by"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

func (Supplier) CollectionName() string {
	return "supplier_masters"
}
