package models

type Unit struct {
	UnitCode int    `json:"unit_code" bson:"unit_code"`
	Name     string `json:"name" bson:"name"`
	Status   bool   `json:"status" bson:"status"`
}

func (Unit) CollectionName() string {
	return "unit"
}
