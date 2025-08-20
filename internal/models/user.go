package models

import "time"

const (
	RoleAdministrator    = "Administrator"
	RolePodft            = "Podft"
	RoleClientManagement = "ClientManagement"
)

type AppUser struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Role      string `gorm:"type:text;not null"`
	Token     string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
}

func (AppUser) TableName() string {
	return "core.users"
}
