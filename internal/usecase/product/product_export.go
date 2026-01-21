package product

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"kakaidee-backend/internal/models"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *ProductService) Export(
	ctx context.Context,
	keyword string,
	startDate string,
	endDate string,
	now time.Time,
) ([]bson.M, error) {

	pipeline := mongo.Pipeline{}

	// ===== keyword filter =====
	if keyword != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"$or": []bson.M{
					{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
					{"sku_code": bson.M{"$regex": keyword, "$options": "i"}},
					{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
				},
			}},
		})
	}

	// ===== date filter =====
	if startDate != "" || endDate != "" {
		dateCond := bson.M{}

		if startDate != "" {
			start, _ := time.Parse("2006-01-02", startDate)
			dateCond["$gte"] = start
		}
		if endDate != "" {
			end, _ := time.Parse("2006-01-02", endDate)
			dateCond["$lte"] = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}

		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"created_at": dateCond,
			}},
		})
	}

	// ===== lookup master =====
	pipeline = append(pipeline,
		bson.D{{"$lookup", bson.M{
			"from":         "category_masters",
			"localField":   "category_code",
			"foreignField": "category_code",
			"as":           "category",
		}}},
		bson.D{{"$lookup", bson.M{
			"from":         "supplier_masters",
			"localField":   "supplier_code",
			"foreignField": "supplier_code",
			"as":           "supplier",
		}}},
		bson.D{{"$lookup", bson.M{
			"from":         "brand_masters",
			"localField":   "brand_code",
			"foreignField": "brand_code",
			"as":           "brand",
		}}},
	)

	// ===== unwind =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
	)

	// ===== project =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"barcode", 1},
			{"sku_code", 1},
			{"product_name", 1},
			{"product_description", 1},
			{"category_code", 1},
			{"category_name", "$category.category_name"},
			{"supplier_code", 1},
			{"supplier_name", "$supplier.supplier_name"},
			{"brand_code", 1},
			{"brand_name", "$brand.brand_name"},
			{"balance_qty", 1},
			{"unit", 1},
			{"cost_price", 1},
			{"created_at", 1},
		}},
	})

	cur, err := s.db.
		Collection(models.Product{}.CollectionName()).
		Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []bson.M
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	end := time.Now().In(loc)
	logProduct := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "product?keyword=" + keyword,
		FunctionMethod:     "GET",
		FunctionName:       "ExportProduct",
		FunctionController: "Product",
		Environment:        "local",
		QueryCollection:    "product_master",
		QueryType:          "query",
		StartTime:          now,
		EndTime:            end,
		DurationMs:         end.Sub(now).Milliseconds(),
		CountData:          len(result),
		StatusCode:         200,
		StatusMessage:      "success",
		CreatedBy:          "admin",
		CreatedAt:          now,
	}
	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	return result, nil
}

func (s *ProductService) ExportExcel(
	ctx context.Context,
	keyword string,
	skuCode string,
	categoryCode string,
	warehouseName string,
	lotsNo string,
	status string,
	now time.Time,
) (string, string, error) {

	endpoint := "product/search-v2?keyword=" + keyword + "&sku_code=" + skuCode + "&category_code=" + categoryCode + "&warehouse_name=" + warehouseName + "&lot_no=" + lotsNo + "&status=" + status
	pipeline := mongo.Pipeline{}

	// ===== 1. match product_master =====
	match := bson.M{}

	if keyword != "" {
		match["$or"] = []bson.M{
			{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
			{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

	if skuCode != "" {
		match["sku_code"] = skuCode
	}

	if categoryCode != "" {
		match["category_code"] = ToInt32(categoryCode)
	}

	if status != "" {
		match["status"] = status
	}

	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", match}})
	}

	// ===== 2. lookup product_stock =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "product_stock",
		"localField":   "barcode",
		"foreignField": "barcode",
		"as":           "stock",
	}}})

	// ===== 3. unwind stock =====
	pipeline = append(pipeline, bson.D{{"$unwind", bson.M{
		"path":                       "$stock",
		"preserveNullAndEmptyArrays": true,
	}}})

	// ===== 4. filter stock =====
	stockMatch := bson.M{}

	if warehouseName != "" {
		stockMatch["stock.warehouses_name"] = warehouseName
	}

	if lotsNo != "" {
		stockMatch["stock.lots_no"] = lotsNo
	}

	if len(stockMatch) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", stockMatch}})
	}

	// ===== 5. lookup category =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "category_masters",
		"localField":   "category_code",
		"foreignField": "category_code",
		"as":           "category",
	}}})

	// ===== 6. lookup supplier =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "supplier_masters",
		"localField":   "supplier_code",
		"foreignField": "supplier_code",
		"as":           "supplier",
	}}})

	// ===== 7. lookup brand =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "brand_masters",
		"localField":   "brand_code",
		"foreignField": "brand_code",
		"as":           "brand",
	}}})

	// ===== 8. lookup unit =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "unit",
		"localField":   "unit",      // product_master.unit
		"foreignField": "unit_code", // unit.unit_code
		"as":           "unit_info",
	}}})

	// ===== 9. unwind master =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$unit_info", "preserveNullAndEmptyArrays": true}}},
	)

	// ===== 10. project (RESULT FINAL) =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"_id": 0,

			"barcode":             1,
			"sku_code":            1,
			"product_name":        1,
			"product_description": 1,

			"brand_code": 1,
			"brand_name": "$brand.brand_name",

			"category_code":    1,
			"category_name_th": "$category.category_name_th",
			"category_name_en": "$category.category_name_en",

			"supplier_code": 1,
			"supplier_name": "$supplier.supplier_name",

			"cost_price":  1,
			"balance_qty": 1,
			"unit":        "$unit_info.unit_code",
			"unit_name":   "$unit_info.name",

			"warehouse_name": "$stock.warehouses_name",
			"warehouse_zone": "$stock.warehouses_zone",
			"bin":            "$stock.bin",
			"stock_type":     "$stock.stock_type",
			"lot_no":         "$stock.lots_no",
			"mfg":            "$stock.mfg",
			"exp":            "$stock.exp",

			"status":     1,
			"created_at": 1,
			"created_by": 1,
			"updated_at": 1,
			"updated_by": 1,
		}},
	})

	cur, err := s.db.
		Collection(models.Product{}.CollectionName()).
		Aggregate(ctx, pipeline)
	if err != nil {

		loc, _ := time.LoadLocation("Asia/Bangkok")
		end := time.Now().In(loc)
		logProduct := models.TransactionLog{
			RequestID:          "",
			FunctionEndpoint:   endpoint,
			FunctionMethod:     "GET",
			FunctionName:       "SearchProduct",
			FunctionController: "Product",
			Environment:        "local",
			QueryCollection:    "product_master",
			QueryType:          "query",
			StartTime:          now,
			EndTime:            end,
			DurationMs:         end.Sub(now).Milliseconds(),
			CountData:          0,
			StatusCode:         400,
			StatusMessage:      "Aggregate fali",
			CreatedBy:          "admin",
			CreatedAt:          now,
		}
		_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

		return "", "", err
	}
	defer cur.Close(ctx)

	var data []bson.M
	if err := cur.All(ctx, &data); err != nil {

		loc, _ := time.LoadLocation("Asia/Bangkok")
		end := time.Now().In(loc)
		logProduct := models.TransactionLog{
			RequestID:          "",
			FunctionEndpoint:   endpoint,
			FunctionMethod:     "GET",
			FunctionName:       "SearchProduct",
			FunctionController: "Product",
			Environment:        "local",
			QueryCollection:    "product_master",
			QueryType:          "query",
			StartTime:          now,
			EndTime:            end,
			DurationMs:         end.Sub(now).Milliseconds(),
			CountData:          len(data),
			StatusCode:         400,
			StatusMessage:      "Result fali",
			CreatedBy:          "admin",
			CreatedAt:          now,
		}
		_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

		return "", "", err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	end := time.Now().In(loc)
	logProduct := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   endpoint,
		FunctionMethod:     "GET",
		FunctionName:       "SearchProduct",
		FunctionController: "Product",
		Environment:        "local",
		QueryCollection:    "product_master",
		QueryType:          "query",
		StartTime:          now,
		EndTime:            end,
		DurationMs:         end.Sub(now).Milliseconds(),
		CountData:          len(data),
		StatusCode:         200,
		StatusMessage:      "success",
		CreatedBy:          "admin",
		CreatedAt:          now,
	}
	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	// ===== header =====
	headers := []string{
		"Barcode", "SKU Code", "Product Name", "Product Description",
		"Brand Code", "Brand Name", "Category Code", "Category Name TH", "Category Name EN",
		"Supplier Code", "Supplier Name",
		"Cost Price", "Balance Qty", "Unit Code", "Unit Name",
		"Warehouse Name", "Warehouse Zone", "Bin", "Stock Type",
		"Lot No", "MFG", "EXP", "Status",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// ===== data =====
	for r, row := range data {
		rowNum := r + 2

		f.SetCellValue(sheet, "A"+strconv.Itoa(rowNum), row["barcode"])
		f.SetCellValue(sheet, "B"+strconv.Itoa(rowNum), row["sku_code"])
		f.SetCellValue(sheet, "C"+strconv.Itoa(rowNum), row["product_name"])
		f.SetCellValue(sheet, "D"+strconv.Itoa(rowNum), row["product_description"])
		f.SetCellValue(sheet, "E"+strconv.Itoa(rowNum), row["brand_code"])
		f.SetCellValue(sheet, "F"+strconv.Itoa(rowNum), row["brand_name"])
		f.SetCellValue(sheet, "G"+strconv.Itoa(rowNum), row["category_code"])
		f.SetCellValue(sheet, "H"+strconv.Itoa(rowNum), row["category_name_th"])
		f.SetCellValue(sheet, "I"+strconv.Itoa(rowNum), row["category_name_en"])
		f.SetCellValue(sheet, "J"+strconv.Itoa(rowNum), row["supplier_code"])
		f.SetCellValue(sheet, "K"+strconv.Itoa(rowNum), row["supplier_name"])
		f.SetCellValue(sheet, "L"+strconv.Itoa(rowNum), row["cost_price"])
		f.SetCellValue(sheet, "M"+strconv.Itoa(rowNum), row["balance_qty"])
		f.SetCellValue(sheet, "N"+strconv.Itoa(rowNum), row["unit_code"])
		f.SetCellValue(sheet, "O"+strconv.Itoa(rowNum), row["unit_name"])
		f.SetCellValue(sheet, "P"+strconv.Itoa(rowNum), row["warehouse_name"])
		f.SetCellValue(sheet, "Q"+strconv.Itoa(rowNum), row["warehouse_zone"])
		f.SetCellValue(sheet, "R"+strconv.Itoa(rowNum), row["bin"])
		f.SetCellValue(sheet, "S"+strconv.Itoa(rowNum), row["stock_type"])
		f.SetCellValue(sheet, "T"+strconv.Itoa(rowNum), row["lot_no"])
		f.SetCellValue(sheet, "U"+strconv.Itoa(rowNum), row["mfg"])
		f.SetCellValue(sheet, "V"+strconv.Itoa(rowNum), row["exp"])
		f.SetCellValue(sheet, "W"+strconv.Itoa(rowNum), row["status"])
	}

	// ===== encode base64 =====
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", "", err
	}

	base64File := base64.StdEncoding.EncodeToString(buf.Bytes())

	fileName := fmt.Sprintf(
		"product_export_%s.xlsx",
		time.Now().Format("20060102_150405"),
	)

	return fileName, base64File, nil
}
