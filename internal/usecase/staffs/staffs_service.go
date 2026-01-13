package staffs

import (
	"context"
	"time"

	"kakaidee-backend/internal/models"
	staffErr "kakaidee-backend/internal/usecase/error/staffs"
)

type StaffsService struct {
	// repo จะใส่ทีหลัง
}

func NewStaffsService() *StaffsService {
	return &StaffsService{}
}

func (s *StaffsService) Create(ctx context.Context, staff models.Staff) error {
	if staff.StaffCode == "" {
		return staffErr.ErrInvalidStaff
	}

	staff.Status = "active"
	staff.CreatedAt = time.Now()
	staff.UpdatedAt = time.Now()

	return nil
}
