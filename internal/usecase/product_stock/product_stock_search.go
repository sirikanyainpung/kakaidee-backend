package product_stock

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
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
) ([]string, error) {

	col := s.db.Collection(models.ProductStock{}.CollectionName())

	// ===== case 1: lot_no มีค่า =====
	if lotNo != "" {
		filter := bson.M{"lot_no": lotNo}

		result, err := col.Distinct(ctx, "lot_no", filter)
		if err != nil {
			return nil, err
		}

		lotList := make([]string, 0, len(result))
		for _, v := range result {
			if s, ok := v.(string); ok {
				lotList = append(lotList, s)
			}
		}

		return lotList, nil
	}

	// ===== case 2: lot_no ว่าง → เอาทั้งหมด =====
	result, err := col.Distinct(ctx, "lot_no", bson.M{})
	if err != nil {
		return nil, err
	}

	lotList := make([]string, 0, len(result))
	for _, v := range result {
		if s, ok := v.(string); ok {
			lotList = append(lotList, s)
		}
	}

	return lotList, nil
}

// ===== search by warehouse_name =====
func (s *ProductStockService) GetByWarehouseName(
	ctx context.Context,
	warehouseName string,
) ([]models.ProductStock, error) {

	filter := bson.M{
		"warehouses_name": warehouseName,
	}

	cur, err := s.db.
		Collection(models.ProductStock{}.CollectionName()).
		Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []models.ProductStock
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}
