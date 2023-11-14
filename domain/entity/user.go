// ⚠️ Don't forget to Add Your new Table
// at AllTables func at all_entities.go file

package entity

import (
	"time"

	"github.com/Lukman-e/try-gost/domain/base"
)

type User struct {
	ID               int        `gorm:"type:serial;primaryKey" json:"id"`
	Name             string     `gorm:"type:varchar(100) not null" json:"name"`
	Email            string     `gorm:"type:varchar(100) not null unique" json:"email"`
	Password         string     `gorm:"type:varchar(255) not null" json:"password"`
	VerificationCode *string    `gorm:"type:varchar(100) null" json:"verification_code"`
	ActivatedAt      *time.Time `gorm:"type:timestamp null;default:null" json:"activated_at"`
	Roles            []Role     `gorm:"many2many:user_has_roles" json:"roles"`
	base.TimeFields
}

func (e *User) SetActivateAccount() {
	timeNow := time.Now()
	e.ActivatedAt = &timeNow
	e.VerificationCode = nil
}

func (e *User) TableName() string {
	return "users"
}

type UserHasRoles struct {
	UserID int  `json:"role_id"`
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
	RoleID int  `json:"permission_id"`
	Role   Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"role"`
}

func (e *UserHasRoles) TableName() string {
	return "user_has_roles"
}
