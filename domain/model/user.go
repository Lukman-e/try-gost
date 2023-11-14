package model

import "github.com/Lukman-e/try-gost/domain/entity"

type UserRegister struct {
	Name     string `validate:"required,min=2,max=60" json:"name"`
	Email    string `validate:"required,email,min=5,max=60" json:"email"`
	Password string `validate:"required,min=8,max=30" json:"password"`
	RoleID   int    `validate:"required,numeric,min=1" json:"role_id"`
}

type UserVerificationCode struct {
	Code  string `validate:"required,min=21,max=60" json:"code"`
	Email string `validate:"required,email,min=5,max=60" json:"email"`
}

type UserLogin struct {
	Email    string `validate:"required,email,min=5,max=60" json:"email"`
	Password string `validate:"required,min=8,max=30" json:"password"`
	IP       string `validate:"required,min=4,max=20" json:"ip"`
}

type UserForgetPassword struct {
	Email string `validate:"required,email,min=5,max=60" json:"email"`
}

type UserResetPassword struct {
	Email              string `validate:"required,email,min=5,max=60" json:"email"`
	Code               string `validate:"required,min=21,max=60" json:"code"`
	NewPassword        string `validate:"required,min=8,max=30" json:"new_password"`
	NewPasswordConfirm string `validate:"required,min=8,max=30" json:"new_password_confirm"`
}

type UserPasswordUpdate struct {
	ID                 int    `validate:"required,numeric,min=1"`
	OldPassword        string `validate:"required,min=8,max=30" json:"old_password"`
	NewPassword        string `validate:"required,min=8,max=30" json:"new_password"`
	NewPasswordConfirm string `validate:"required,min=8,max=30" json:"new_password_confirm"`
}

type UserProfile struct {
	Email string
	Name  string
	Role  entity.Role
}
