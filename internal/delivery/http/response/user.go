package response

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email" `
	Phone string `json:"phone"`
	Role  string `json:"role"`
}
