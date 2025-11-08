package dtos

type CreateUserDTO struct {
	Username string `binding:"required" json:"username"`
	Password string `binding:"required,min=6,max=64" json:"password"`
	Email    string `binding:"required,email" json:"email"`
}
