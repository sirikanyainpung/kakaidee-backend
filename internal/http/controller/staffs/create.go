package staffs

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"kakaidee-backend/internal/models"
	payloadStaffs "kakaidee-backend/internal/payload/staffs"
	staffErr "kakaidee-backend/internal/usecase/error/staffs"
	"kakaidee-backend/internal/usecase/staffs"
)

type StaffsController struct {
	svc *staffs.StaffsService
}

func NewStaffsController(svc *staffs.StaffsService) *StaffsController {
	return &StaffsController{svc: svc}
}

func (c *StaffsController) Create(ctx echo.Context) error {
	var req payloadStaffs.CreateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	staff := models.Staff{
		StaffCode: req.StaffCode,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		PhoneNo:   req.PhoneNo,
		Email:     req.Email,
		Address:   req.Address,
	}

	err := c.svc.Create(ctx.Request().Context(), staff)
	if err != nil {
		switch err {
		case staffErr.ErrInvalidStaff:
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		case staffErr.ErrDuplicateStaff:
			return ctx.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
		}
	}

	return ctx.JSON(http.StatusCreated, map[string]string{
		"status": "created",
	})
}
