package product

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"kakaidee-backend/internal/models"
	payloadProduct "kakaidee-backend/internal/payload/product"
)

func (s *ProductService) ImportFromExcel(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	createdBy string,
	now time.Time,
) (map[string]interface{}, error) {

	if s.db == nil {
		return nil, errors.New("database not initialized")
	}

	// ===== open file =====
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}

	rows, err := f.GetRows("Product")
	if err != nil {
		return nil, err
	}

	var (
		success int
		failed  []string
	)

	// ===== parse rows =====
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}

		// ต้องมีอย่างน้อย 17 columns (0-16)
		if len(row) < 17 {
			failed = append(failed,
				fmt.Sprintf("row %d: invalid column length (%d)", i+1, len(row)),
			)
			continue
		}

		req := payloadProduct.ImportProductRequest{
			Barcode:            strings.TrimSpace(row[0]),
			SKUCode:            strings.TrimSpace(row[1]),
			ProductName:        strings.TrimSpace(row[2]),
			ProductDescription: strings.TrimSpace(row[3]),

			CategoryCode: ToInt32(row[4]),
			SupplierCode: ToInt32(row[5]),
			BrandCode:    ToInt32(row[6]),

			Unit:          strings.TrimSpace(row[7]),
			CostPrice:     ToFloat64(row[8]),
			LotNo:         strings.TrimSpace(row[9]),
			WarehouseName: strings.TrimSpace(row[10]),
			WarehouseZone: strings.TrimSpace(row[11]),
			Bin:           strings.TrimSpace(row[12]),
			StockType:     strings.TrimSpace(row[13]),

			ReceiveQty: ToInt(row[14]),
			MFG:        ToDate(row[15]),
			EXP:        ToDate(row[16]),
		}

		// ===== insert product_stock =====
		productStock := models.ProductStock{
			Barcode:       req.Barcode,
			SKUCode:       req.SKUCode,
			LotNo:         req.LotNo,
			WarehouseName: req.WarehouseName,
			WarehouseZone: req.WarehouseZone,
			Bin:           req.Bin,
			StockType:     req.StockType,
			ReceiveQty:    req.ReceiveQty,
			BalanceQty:    req.ReceiveQty,
			SellingQty:    0,
			MFG:           req.MFG,
			EXP:           req.EXP,
			Status:        "active",
			CreatedBy:     createdBy,
			UpdatedBy:     createdBy,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if _, err := s.db.
			Collection(models.ProductStock{}.CollectionName()).
			InsertOne(ctx, productStock); err != nil {

			failed = append(failed,
				fmt.Sprintf("row %d: insert product_stock failed", i+1),
			)
			continue
		}

		// // ===== log product_stock =====
		// _, _ = s.db.
		// 	Collection(models.TransactionLog{}.CollectionName()).
		// 	InsertOne(ctx, buildLog(
		// 		"product/import/excel",
		// 		"product_stock",
		// 		"insert",
		// 		now,
		// 		createdBy,
		// 	))

		// ===== upsert product_master =====
		product := models.Product{
			Barcode:            req.Barcode,
			SKUCode:            req.SKUCode,
			ProductName:        req.ProductName,
			ProductDescription: req.ProductDescription,
			CategoryCode:       req.CategoryCode,
			SupplierCode:       req.SupplierCode,
			BrandCode:          req.BrandCode,
			Unit:               req.Unit,
			CostPrice:          req.CostPrice,
			Status:             "active",
			UpdatedBy:          createdBy,
			UpdatedAt:          now,
		}

		filter := bson.M{"sku_code": req.SKUCode}
		update := bson.M{
			"$set":         product,
			"$inc":         bson.M{"balance_qty": req.ReceiveQty},
			"$setOnInsert": bson.M{"created_at": now, "created_by": createdBy},
		}

		if _, err := s.db.
			Collection(models.Product{}.CollectionName()).
			UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {

			// insert
			if _, err := s.db.
				Collection(models.Product{}.CollectionName()).
				InsertOne(ctx, product); err != nil {
				failed = append(failed,
					fmt.Sprintf("row %d: Insert product failed", i+1),
				)
				continue
			}
			// failed = append(failed,
			// 	fmt.Sprintf("row %d: upsert product failed", i+1),
			// )
			// continue
		}

		success++
	}

	// ===== log product_master =====
	_, _ = s.db.
		Collection(models.TransactionLog{}.CollectionName()).
		InsertOne(ctx, buildLog(
			"product/import/excel",
			"product_master",
			"upsert",
			now,
			createdBy,
			success*2,
		))

	return map[string]interface{}{
		"imported": success,
		"failed":   failed,
	}, nil
}

/* ===================== Helper ===================== */

func ToInt32(val string) int32 {
	v := strings.TrimSpace(val)
	i, _ := strconv.Atoi(v)
	return int32(i)
}

func ToInt(val string) int {
	v := strings.TrimSpace(val)
	i, _ := strconv.Atoi(v)
	return i
}

func ToFloat64(val string) float64 {
	v := strings.TrimSpace(val)
	v = strings.ReplaceAll(v, ",", "")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func ToDate(val string) time.Time {
	v := strings.TrimSpace(val)
	if v == "" {
		return time.Time{}
	}

	// yyyy-mm-dd
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t
	}

	// excel serial date
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		t, _ := excelize.ExcelDateToTime(f, false)
		return t
	}

	return time.Time{}
}

func buildLog(
	endpoint string,
	collection string,
	queryType string,
	now time.Time,
	createdBy string,
	countData int,
) models.TransactionLog {

	return models.TransactionLog{
		FunctionEndpoint:   endpoint,
		FunctionMethod:     "POST",
		FunctionName:       "ImportProductExcel",
		FunctionController: "Product",
		Environment:        "local",
		QueryCollection:    collection,
		QueryType:          queryType,
		StartTime:          now,
		EndTime:            time.Now(),
		DurationMs:         time.Since(now).Milliseconds(),
		CountData:          countData,
		StatusCode:         201,
		StatusMessage:      "created",
		CreatedBy:          createdBy,
		CreatedAt:          time.Now(),
	}
}
