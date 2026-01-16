package models

import "time"

type Brand struct {
	BrandCode   int32     `json:"brand_code" bson:"brand_code"`
	BrandName   string    `json:"brand_name" bson:"brand_name"`
	Status      string    `json:"status" bson:"status"`
	Createbsony string    `json:"created_by" bson:"created_by"`
	Updatebsony string    `json:"updated_by" bson:"updated_by"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

func (Brand) CollectionName() string {
	return "brand_masters"
}
