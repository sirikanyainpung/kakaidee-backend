package product_stock

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
	payloadProduct "kakaidee-backend/internal/payload/product"
)

type ProductStockService struct {
	db *mongo.Database
}

func NewProductStockService(db *mongo.Database) *ProductStockService {
	return &ProductStockService{db: db}
}

// ===== search by lot_no =====
func (s *ProductStockService) GetLotNoList(
	ctx context.Context,
	lotNo string,
) ([]payloadProduct.LotsListResponse, error) {

	col := s.db.Collection(models.ProductStock{}.CollectionName())
	var lotList []payloadProduct.LotsListResponse
	// ===== case 1: lot_no มีค่า =====
	if lotNo != "" {
		filter := bson.M{"lots_no": lotNo}

		result, err := col.Distinct(ctx, "lots_no", filter)
		if err != nil {
			return nil, err
		}

		for _, v := range result {
			if s, ok := v.(string); ok {
				if s != "" {
					lotList = append(lotList, payloadProduct.LotsListResponse{
						LotsNo: s,
					})
				}
			}
		}

		return lotList, nil
	}

	// ===== case 2: lot_no ว่าง → เอาทั้งหมด =====
	result, err := col.Distinct(ctx, "lots_no", bson.M{})
	if err != nil {
		return nil, err
	}

	for _, v := range result {
		if s, ok := v.(string); ok {
			if s != "" {
				lotList = append(lotList, payloadProduct.LotsListResponse{
					LotsNo: s,
				})
			}
		}
	}

	return lotList, nil
}

// ===== search by warehouse_name =====
func (s *ProductStockService) GetByWarehouseName(
	ctx context.Context,
	warehouseName string,
) ([]payloadProduct.WMSListResponse, error) {

	col := s.db.Collection(models.ProductStock{}.CollectionName())
	var wmsList []payloadProduct.WMSListResponse

	// ===== case 1: warehouses_name มีค่า =====
	if warehouseName != "" {
		filter := bson.M{"warehouses_name": warehouseName}

		result, err := col.Distinct(ctx, "warehouses_name", filter)
		if err != nil {
			return nil, err
		}

		for _, v := range result {
			if s, ok := v.(string); ok {
				if s != "" {
					wmsList = append(wmsList, payloadProduct.WMSListResponse{
						WmsName: s,
					})
				}
			}
		}

		return wmsList, nil
	}

	// ===== case 2: warehouses_name ว่าง → เอาทั้งหมด =====
	result, err := col.Distinct(ctx, "warehouses_name", bson.M{})
	if err != nil {
		return nil, err
	}

	for _, v := range result {
		if s, ok := v.(string); ok {
			if s != "" {
				wmsList = append(wmsList, payloadProduct.WMSListResponse{
					WmsName: s,
				})
			}
		}
	}

	return wmsList, nil
}
