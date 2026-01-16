package models

import "time"

type Category struct {
	CategoryCode int32     `json:"category_code" bson:"category_code"`
	CategoryName string    `json:"category_name" bson:"category_name"`
	Status       string    `json:"status" bson:"status"`
	Createbsony  string    `json:"created_by" bson:"created_by"`
	Updatebsony  string    `json:"updated_by" bson:"updated_by"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

func (Category) CollectionName() string {
	return "category_masters"
}
