package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
	RoleSuperAdmin Role = "superadmin"
)

type UserStatus string

const (
	StatusPending  UserStatus = "pending"
	StatusApproved UserStatus = "approved"
	StatusBlocked  UserStatus = "blocked"
)

type User struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Username  string         `json:"username"`
	Role      Role           `gorm:"type:varchar(20);default:'user'" json:"role"`
	Status    UserStatus     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
