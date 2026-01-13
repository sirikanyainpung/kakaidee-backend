package models

import "time"

type Staff struct {
	StaffCode string    `json:"staff_code" bson:"staff_code"`
	FirstName string    `json:"first_name" bson:"staff_firstname"`
	LastName  string    `json:"last_name" bson:"staff_lastname"`
	PhoneNo   string    `json:"phone_no" bson:"phone_no"`
	Email     string    `json:"email" bson:"email"`
	Address   string    `json:"address" bson:"address"`
	Status    string    `json:"status" bson:"status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func (Staff) CollectionName() string {
	return "staffs"
}
