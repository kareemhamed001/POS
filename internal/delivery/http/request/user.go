package request

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,egyptianphone"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name" validate:"omitempty,min=2,max=50"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Phone    *string `json:"phone" validate:"omitempty,egyptianphone"`
	Password *string `json:"password" validate:"omitempty"`
}
