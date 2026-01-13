package staffs

import "errors"

var (
	ErrStaffNotFound  = errors.New("staff not found")
	ErrInvalidStaff   = errors.New("invalid staff")
	ErrDuplicateStaff = errors.New("duplicate staff")
)
