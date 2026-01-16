package product

import (
	"kakaidee-backend/internal/helper"
	"net/http"
	"strconv"
	"time"

	"bytes"
	"encoding/base64"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

func (c *ProductController) Export(ctx echo.Context) error {

	keyword := ctx.QueryParam("keyword")
	startDate := ctx.QueryParam("start_date")
	endDate := ctx.QueryParam("end_date")

	now := time.Now()
	data, err := c.svc.Export(
		ctx.Request().Context(),
		keyword,
		startDate,
		endDate,
		now,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Export failed."),
		)
	}

	// helper.PrintStructJson(" ========== data ==========")
	// helper.PrintStructJson(data)

	// ===== Create Excel in memory =====
	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Barcode", "SKU Code", "Product Name", "Product Description",
		"Category Code", "Category Name",
		"Supplier Code", "Supplier Name",
		"Brand Code", "Brand Name",
		"Balance Qty", "Unit", "Cost Price",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for r, item := range data {
		row := r + 2
		f.SetCellValue(sheet, "A"+strconv.Itoa(row), item["barcode"])
		f.SetCellValue(sheet, "B"+strconv.Itoa(row), item["sku_code"])
		f.SetCellValue(sheet, "C"+strconv.Itoa(row), item["product_name"])
		f.SetCellValue(sheet, "D"+strconv.Itoa(row), item["product_description"])
		f.SetCellValue(sheet, "E"+strconv.Itoa(row), item["category_code"])
		f.SetCellValue(sheet, "F"+strconv.Itoa(row), item["category_name"])
		f.SetCellValue(sheet, "G"+strconv.Itoa(row), item["supplier_code"])
		f.SetCellValue(sheet, "H"+strconv.Itoa(row), item["supplier_name"])
		f.SetCellValue(sheet, "I"+strconv.Itoa(row), item["brand_code"])
		f.SetCellValue(sheet, "J"+strconv.Itoa(row), item["brand_name"])
		f.SetCellValue(sheet, "K"+strconv.Itoa(row), item["balance_qty"])
		f.SetCellValue(sheet, "L"+strconv.Itoa(row), item["unit"])
		f.SetCellValue(sheet, "M"+strconv.Itoa(row), item["cost_price"])
	}

	// ===== Write to buffer =====
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", "Export product failed."),
		)
	}

	// ===== Encode base64 =====
	base64File := base64.StdEncoding.EncodeToString(buf.Bytes())

	type ExportResponse struct {
		ExportName string `json:"export_name"`
		Base64File string `json:"base64_file"`
	}

	result := ExportResponse{
		ExportName: "products_" + now.Format("20060102_150405") + ".xlsx",
		Base64File: base64File,
	}

	return ctx.JSON(
		http.StatusOK,
		helper.Success("Export product success.", result),
	)
}
