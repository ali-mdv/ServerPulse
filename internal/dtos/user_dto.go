package dtos

type CreateUserDTO struct {
	Email    string `binding:"required,email" json:"email"`
	Password string `binding:"required,min=6,max=64" json:"password"`
}

type UpdateUserDTO struct {
	Email    *string `binding:"omitempty,email" json:"email"`
	Password *string `binding:"omitempty,min=6,max=64" json:"password"`
}
