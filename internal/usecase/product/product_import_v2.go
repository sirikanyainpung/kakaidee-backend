package product

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/models"
	payloadProduct "kakaidee-backend/internal/payload/product"
)

func (s *ProductService) ImportFromExcelV2(
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
		successStock         int
		failedStock          int
		success              int
		successProductUpdate int
		failedProduct        int
		failed               []string
		productListInsert    []interface{}
		productStockInsert   []interface{}
	)

	// เช็คข้อมูลใน DB
	// กำหนด context สำหรับการเชื่อมต่อ
	// ctx, cancel := context.WithTimeout(context.Background(), 900*time.Second)
	// defer cancel()

	// สร้าง query filter หากต้องการกรองข้อมูล (ตัวอย่าง: แสดงทั้งหมด)
	filter := bson.M{} // หรือใส่ filter ที่ต้องการกรอง เช่น {"status": "active"}

	// สร้าง projection เพื่อดึงเฉพาะฟิลด์ barcode
	projection := bson.M{
		"barcode": 1, // เลือกฟิลด์ barcode เท่านั้น
		"_id":     0, // ไม่ดึง _id
	}

	// query product_master collection
	cur, err := s.db.
		Collection("product_masters").                              // ชื่อ collection
		Find(ctx, filter, options.Find().SetProjection(projection)) // ใช้ projection
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	// สร้าง map สำหรับเก็บผลลัพธ์
	productMap := make(map[string]interface{}) // ใช้ map สำหรับการค้นหาด้วย barcode

	// ดึงข้อมูลทั้งหมดจาก cursor
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			return nil, err
		}

		// ใช้ barcode เป็น key และเก็บข้อมูลใน map
		if barcode, ok := result["barcode"].(string); ok {
			productMap[barcode] = result["barcode"]
		}
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	// แสดงผลลัพธ์
	// helper.PrintStructJson(" ----- productMap ----- ")
	// helper.PrintStructJson(productMap)

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

			Unit:          ToInt32(row[7]),
			CostPrice:     ToFloat64(row[8]),
			LotNo:         strings.TrimSpace(row[9]),
			WarehouseName: strings.TrimSpace(row[10]),
			WarehouseZone: strings.TrimSpace(row[11]),
			Bin:           strings.TrimSpace(row[12]),
			StockType:     strings.TrimSpace(row[13]),

			ReceiveQty: ToInt32(row[14]),
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
			Status:        true,
			CreatedBy:     createdBy,
			UpdatedBy:     createdBy,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		productStockInsert = append(productStockInsert, productStock)

		// ตรวจสอบว่าเก็บข้อมูลทั้งหมดใน productStocks แล้วทำการ insert ทีเดียว
		if len(productStockInsert) >= 1000 {
			helper.PrintStructJson(" ----- productStockInsert ----- ")
			// helper.PrintStructJson(productStockInsert)
			// ใช้ InsertMany เพื่อเพิ่มข้อมูลหลายแถว
			_, err := s.db.
				Collection(models.ProductStock{}.CollectionName()).
				InsertMany(ctx, productStockInsert)
			if err != nil {
				failed = append(failed,
					fmt.Sprintf("สินค้า %d รายการ: Insert product-stock failed: %v", len(productStockInsert), err),
				)

				failedStock = failedStock + len(productStockInsert)
				helper.PrintStructJson(" -----product-stock fail ----- ")
			} else {
				successStock = successStock + len(productStockInsert)
				helper.PrintStructJson(" -----product-stock success ----- ")
			}

			// เคลียร์ productStocks เพื่อเตรียมข้อมูลสำหรับการ insert รอบถัดไป
			productStockInsert = []interface{}{}
		}

		// ตัวอย่างการค้นหาข้อมูลจาก map
		if _, exists := productMap[req.Barcode]; exists {
			// helper.PrintStructJson(" ----- Update ----- ")

			filter := bson.M{"barcode": req.Barcode}
			update := bson.M{
				"$set": bson.M{
					"balance_qty": req.ReceiveQty,
					"updated_by":  createdBy,
					"updated_at":  now,
				},
			}

			if _, err := s.db.
				Collection(models.Product{}.CollectionName()).
				UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
				failed = append(failed,
					fmt.Sprintf("แถว %d: Update product failed :: %v", i+1, err),
				)
				failedProduct++
				helper.PrintStructJson(" -----product update fail ----- ")
			} else {
				successProductUpdate++
				helper.PrintStructJson(" -----product update success ----- ")
			}
		} else {
			// helper.PrintStructJson(" ----- Insert ----- ")

			product := models.Product{
				Barcode:            req.Barcode,
				SKUCode:            req.SKUCode,
				ProductName:        req.ProductName,
				ProductDescription: req.ProductDescription,
				CategoryCode:       req.CategoryCode,
				SupplierCode:       req.SupplierCode,
				BrandCode:          req.BrandCode,
				CostPrice:          req.CostPrice,
				BalanceQty:         req.ReceiveQty,
				Status:             true,
				CreatedBy:          createdBy,
				CreatedAt:          now,
				UpdatedBy:          createdBy,
				UpdatedAt:          now,
			}
			productListInsert = append(productListInsert, product)
		}

		// ตรวจสอบว่าเก็บข้อมูลทั้งหมดใน product แล้วทำการ insert ทีเดียว
		if len(productListInsert) >= 1000 {
			helper.PrintStructJson(" ----- productListInsert ----- ")
			// helper.PrintStructJson(productListInsert)
			// ใช้ InsertMany เพื่อเพิ่มข้อมูลหลายแถว
			_, err := s.db.
				Collection(models.Product{}.CollectionName()).
				InsertMany(ctx, productListInsert)
			if err != nil {
				failed = append(failed,
					fmt.Sprintf("สินค้า %d รายการ: Insert product failed: %v", len(productListInsert), err),
				)
				failedProduct = failedProduct + len(productListInsert)
				helper.PrintStructJson(" -----product insert failed ----- ")
				helper.PrintStructJson(err)
			} else {
				success = success + len(productListInsert)
				helper.PrintStructJson(" -----product insert success ----- ")
			}

			// เคลียร์ productStocks เพื่อเตรียมข้อมูลสำหรับการ insert รอบถัดไป
			productListInsert = []interface{}{}
		}
	}

	if len(productStockInsert) > 0 {
		helper.PrintStructJson(" ----- productStockInsert > 0 ----- ")
		// helper.PrintStructJson(productStockInsert)
		_, err := s.db.
			Collection(models.ProductStock{}.CollectionName()).
			InsertMany(ctx, productStockInsert)
		if err != nil {
			failed = append(failed,
				fmt.Sprintf("สินค้า %d: Insert product_stock failed: %v", len(productStockInsert), err),
			)

			failedStock = failedStock + len(productStockInsert)
		} else {
			successStock = successStock + len(productStockInsert)
		}

		productStockInsert = []interface{}{}
	}

	if len(productListInsert) > 0 {
		helper.PrintStructJson(" ----- productListInsert > 0 ----- ")
		// helper.PrintStructJson(productListInsert)
		_, err := s.db.
			Collection(models.Product{}.CollectionName()).
			InsertMany(ctx, productListInsert)
		if err != nil {
			failed = append(failed,
				fmt.Sprintf("สินค้า %d: Insert product failed: %v", len(productListInsert), err),
			)

			failedProduct = failedProduct + len(productListInsert)
		} else {
			success = success + len(productListInsert)
		}

		productListInsert = []interface{}{}
	}

	// ===== log product_master success =====
	if (success + successStock) > 0 {
		_, _ = s.db.
			Collection(models.TransactionLog{}.CollectionName()).
			InsertOne(ctx, buildLog(
				"product/import/excel-v2",
				"product_master",
				"insert",
				now,
				createdBy,
				success+successStock,
				201,
				"created",
				requestID,
			))
	}

	// ===== log product_master success =====
	if (successProductUpdate) > 0 {
		_, _ = s.db.
			Collection(models.TransactionLog{}.CollectionName()).
			InsertOne(ctx, buildLog(
				"product/import/excel-v2",
				"product_master",
				"upsert",
				now,
				createdBy,
				successProductUpdate,
				200,
				"updated",
				requestID,
			))
	}

	// ===== log product_master false =====
	if (failedProduct + failedStock) > 0 {
		_, _ = s.db.
			Collection(models.TransactionLog{}.CollectionName()).
			InsertOne(ctx, buildLog(
				"product/import/excel-v2",
				"product_master",
				"upsert",
				now,
				createdBy,
				failedProduct+failedStock,
				400,
				"fail",
				requestID,
			))
	}

	resutlMassage := ""
	if successProductUpdate > 0 {
		resutlMassage = "อัปเดตข้อมูลสำเร็จ " + strconv.Itoa(successProductUpdate) + " ข้อมูล"
	}

	if success > 0 {
		if resutlMassage != "" {
			resutlMassage = resutlMassage + " เพิ่มข้อมูลสำเร็จ " + strconv.Itoa(success) + " ข้อมูล"
		}
		resutlMassage = "เพิ่มข้อมูลสำเร็จ " + strconv.Itoa(success) + " ข้อมูล"
	}

	if failedProduct > 0 {
		if resutlMassage != "" {
			resutlMassage = resutlMassage + " เพิ่มข้อมูลไม่สำเร็จ " + strconv.Itoa(failedProduct) + " ข้อมูล"
		}
		resutlMassage = "เพิ่มข้อมูลไม่สำเร็จ " + strconv.Itoa(failedProduct) + " ข้อมูล"
	}

	helper.PrintStructJson(" ---- success --- ")
	helper.PrintStructJson(success)
	helper.PrintStructJson(" ---- successProductUpdate ---- ")
	helper.PrintStructJson(successProductUpdate)
	helper.PrintStructJson(" ---- failedProduct ---- ")
	helper.PrintStructJson(failedProduct)
	helper.PrintStructJson(" ---- failedStock ---- ")
	helper.PrintStructJson(failedStock)

	return map[string]interface{}{
		"imported":     success,
		"failed":       failed,
		"finalMassage": resutlMassage,
	}, nil
}
