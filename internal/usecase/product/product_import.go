package product

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
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

	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(1000000000) + 1
	requestID := "103-" + strconv.Itoa(randomInt)

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
		successStock int
		failedStock  int
		success      int
		failed       []string
		failMassage  string
	)

	// ===== parse rows =====
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}

		// ต้องมีอย่างน้อย 17 columns (0-16)
		if len(row) < 17 {
			failed = append(failed,
				fmt.Sprintf("แถว %d: invalid column length (%d)", i+1, len(row)),
			)

			failMassage = fmt.Sprintf("แถว %d", i+1)
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

			// failed = append(failed,
			// 	fmt.Sprintf("row %d: insert product_stock failed", i+1),
			// )

			failed = append(failed,
				fmt.Sprintf("row %d: Insert product_stock failed: %v", i+1, err),
			)
			if failMassage == "" {
				failMassage = fmt.Sprintf("แถว %d", i+1)
			} else {
				failMassage = failMassage + "," + fmt.Sprintf("แถว %d", i+1)
			}

			failedStock++
			continue
		}

		successStock++

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
			BalanceQty:         req.ReceiveQty,
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
			product.CreatedAt = now
			product.CreatedBy = createdBy
			// insert
			if _, err := s.db.
				Collection(models.Product{}.CollectionName()).
				InsertOne(ctx, product); err != nil {
				// failed = append(failed,
				// 	fmt.Sprintf("row %d: Insert product failed", i+1),
				// )
				failed = append(failed,
					fmt.Sprintf("row %d: Insert product failed: %v", i+1, err),
				)

				if failMassage == "" {
					failMassage = fmt.Sprintf("แถว %d", i+1)
				} else {
					failMassage = failMassage + "," + fmt.Sprintf("แถว %d", i+1)
				}

				continue
			}
			// failed = append(failed,
			// 	fmt.Sprintf("row %d: upsert product failed", i+1),
			// )
			// continue
		}

		success++
	}

	// ===== log product_master success =====
	if (success + successStock) > 0 {
		_, _ = s.db.
			Collection(models.TransactionLog{}.CollectionName()).
			InsertOne(ctx, buildLog(
				"product/import/excel",
				"product_master",
				"upsert",
				now,
				createdBy,
				success+successStock,
				201,
				"created",
				requestID,
			))
	}

	// ===== log product_master false =====
	if (len(failed) + failedStock) > 0 {
		_, _ = s.db.
			Collection(models.TransactionLog{}.CollectionName()).
			InsertOne(ctx, buildLog(
				"product/import/excel",
				"product_master",
				"upsert",
				now,
				createdBy,
				len(failed)+failedStock,
				400,
				"fail",
				requestID,
			))
	}

	resutlMassage := ""
	if failMassage == "" {
		resutlMassage = "เพิ่มข้อมูลสำเร็จ " + strconv.Itoa(success) + " ข้อมูล"
	} else {
		resutlMassage = "เพิ่มข้อมูลสำเร็จ " + strconv.Itoa(success) + " ข้อมูล และเพิ่มข้อมูลไม่สำเร็จ " + strconv.Itoa(len(failed)+failedStock) + " ข้อมูล"
	}

	return map[string]interface{}{
		"imported":     success,
		"failed":       failed,
		"finalMassage": resutlMassage,
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
	statusCode int,
	statusMessage string,
	requestID string,
) models.TransactionLog {

	loc, _ := time.LoadLocation("Asia/Bangkok")

	return models.TransactionLog{
		RequestID:          requestID,
		FunctionEndpoint:   endpoint,
		FunctionMethod:     "POST",
		FunctionName:       "ImportProductExcel",
		FunctionController: "Product",
		Environment:        "local",
		QueryCollection:    collection,
		QueryType:          queryType,
		StartTime:          now,
		EndTime:            time.Now().In(loc),
		DurationMs:         time.Since(now).Milliseconds(),
		CountData:          countData,
		StatusCode:         statusCode,
		StatusMessage:      statusMessage,
		CreatedBy:          createdBy,
		CreatedAt:          time.Now(),
	}
}
