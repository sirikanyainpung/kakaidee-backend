package transactionLog

import (
	"net/http"
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

	result, err := c.svc.GetTransactionLog(
		ctx.Request().Context(),
		now,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError,
			helper.Error(500, "error", err.Error()),
		)
	}

	return ctx.JSON(http.StatusOK,
		helper.Success("Get lots_no list success.", result),
	)
}
