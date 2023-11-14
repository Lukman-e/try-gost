package model

type UserCreate struct {
	Name     string `validate:"required,min=5,max=60" json:"name"`
	Email    string `validate:"required,email,min=5,max=60" json:"email"`
	Password string `validate:"required,min=8,max=30" json:"password"`
	IsAdmin  bool   `validate:"boolean" json:"is_admin"`
}

type UserResponse struct {
	ID    int    `validate:"required,numeric,min=1" json:"id"`
	Name  string `validate:"required,min=5,max=60" json:"name"`
	Email string `validate:"required,email,min=5,max=60" json:"email"`
}

type UserProfileUpdate struct {
	ID   int    `validate:"required,numeric,min=1"`
	Name string `validate:"required,min=5,max=60" json:"name"`
	// ...
	// add more fields
}
