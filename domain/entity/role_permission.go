// ⚠️ Don't forget to Add Your new Table
// at AllTables func at all_entities.go file

package entity

import "github.com/Lukman-e/try-gost/domain/base"

// This vars used in userServiceLayer
const (
	ADMIN = 1
	USER  = 2
	// ...
	// Add your own roleID
)

type Role struct {
	ID          int          `gorm:"type:serial;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(255) not null unique" json:"name"`
	Description string       `gorm:"type:varchar(255) not null" json:"description"`
	Permissions []Permission `gorm:"many2many:role_has_permissions" json:"permissions"`
	base.TimeFields
}

func (e *Role) TableName() string {
	return "roles"
}

type RoleHasPermission struct {
	RoleID       int        `json:"role_id"`
	Role         Role       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"role"`
	PermissionID int        `json:"permission_id"`
	Permission   Permission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"permission"`
}

func (e *RoleHasPermission) TableName() string {
	return "role_has_permissions"
}

type Permission struct {
	ID          int    `gorm:"type:serial;primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(255) not null unique" json:"name"`
	Description string `gorm:"type:varchar(255) not null" json:"description"`
	base.TimeFields
}

func (e *Permission) TableName() string {
	return "permissions"
}
