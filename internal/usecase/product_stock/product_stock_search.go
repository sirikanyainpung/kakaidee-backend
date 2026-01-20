package product_stock

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/helper"
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

	pipeline := mongo.Pipeline{}
	// ===== 1. match (search) =====
	if warehouseName != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"$or": []bson.M{
					{"warehouses_name": bson.M{"$regex": warehouseName, "$options": "i"}},
				},
			}},
		})
	}

	// ===== 2. distinct warehouse_name and warehouse_zone =====
	pipeline = append(pipeline, bson.D{
		{"$group", bson.M{
			"_id": bson.M{
				"warehouses_name": "$warehouses_name", // group by warehouses_name
				"warehouses_zone": "$warehouses_zone", // group by warehouses_zone
			},
		}},
	})

	// ===== 3. project to return the final fields =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"warehouses_name": "$_id.warehouses_name",
			"warehouses_zone": "$_id.warehouses_zone",
			"_id":             0, // remove _id from the result
		}},
	})

	// Execute the aggregation
	cur, err := s.db.Collection(models.ProductStock{}.CollectionName()).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []bson.M
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}

	// Helper print to debug
	helper.PrintStructJson(" ----- result all ----- ")
	helper.PrintStructJson(result)

	var wmsList []payloadProduct.WMSListResponse
	// Loop through the result and populate wmsList
	for _, v := range result {
		// v["warehouse_name"] and v["warehouse_zone"] should contain the values we grouped by
		if warehouseName, ok := v["warehouses_name"].(string); ok {
			if warehouseName != "" {
				if warehouseZone, ok := v["warehouses_zone"].(string); ok {
					wmsList = append(wmsList, payloadProduct.WMSListResponse{
						WmsName: warehouseName,
						WmsZone: warehouseZone,
					})
				}
			}
		}
	}

	return wmsList, nil
}
