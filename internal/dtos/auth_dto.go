package dtos

type LoginDTO struct {
	Password string `binding:"required" json:"password"`
	Email    string `binding:"required,email" json:"email"`
}
