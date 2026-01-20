package product

type LotsListResponse struct {
	LotsNo string `json:"lot_no" bson:"lots_no"`
}

type WMSListResponse struct {
	WmsName string `json:"warehouse_name" bson:"warehouses_name"`
}
