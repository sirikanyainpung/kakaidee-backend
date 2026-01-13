package payload

type CreateStaffRequest struct {
	StaffCode string `json:"staff_code"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhoneNo   string `json:"phone_no"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}
