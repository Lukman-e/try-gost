package model

type RoleCreate struct {
	Name          string `validate:"required,min=1,max=60" json:"name"`
	Description   string `validate:"required,min=1,max=100" json:"description"`
	PermissionsID []int  `json:"permissions_id"` // can null
}

type RoleResponse struct {
	ID          int    `validate:"required,numeric,min=1" json:"id"`
	Name        string `validate:"required,min=1" json:"name"`
	Description string `validate:"required,min=1" json:"description"`
}

type RoleUpdate struct {
	ID          int    `validate:"required,numeric,min=1"`
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1,max=100" json:"description"`
}

type RoleConnectToPermissions struct {
	RoleID        int   `validate:"required,numeric,min=1" json:"role_id"`
	PermissionsID []int `validate:"required" json:"permissions_id"`
}

type PermissionCreate struct {
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1,max=100" json:"description"`
}

type PermissionUpdate struct {
	ID          int    `validate:"required,numeric,min=1"`
	Name        string `validate:"required,min=1,max=60" json:"name"`
	Description string `validate:"required,min=1,max=100" json:"description"`
}

type PermissionResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
