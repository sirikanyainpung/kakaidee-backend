package transactionLog

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"kakaidee-backend/internal/helper"
	transactionLog "kakaidee-backend/internal/usecase/transaction_log"
)

type TransactionLogController struct {
	svc *transactionLog.TransactionLogService
}

func NewTransactionLogController(
	svc *transactionLog.TransactionLogService,
) *TransactionLogController {
	return &TransactionLogController{svc: svc}
}

// ===== GET Transaction Log =====
func (c *TransactionLogController) GetTransactionLog(ctx echo.Context) error {

	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	pageStr := ctx.QueryParam("page")
	limitStr := ctx.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	result, err := c.svc.GetTransactionLog(
		ctx.Request().Context(),
		now,
		page,
		limit,
		skip,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", err.Error()),
		)
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Get transaction log success.", result),
	)
}

// ===== Export Transaction Log =====
func (c *TransactionLogController) ExportTransactionLog(ctx echo.Context) error {

	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	fileName, base64File, err := c.svc.ExportTransactionLog(
		ctx.Request().Context(),
		now,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", err.Error()),
		)
	}

	result := map[string]string{
		"export_name": fileName,
		"base64_file": base64File,
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Export transaction log success.", result),
	)

}
