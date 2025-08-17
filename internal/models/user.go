package models

import "time"

type AppUser struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Role      string `gorm:"type:text;not null"`   // admin | staff | viewer
	Token     string `gorm:"uniqueIndex;not null"` // dev-токен для Bearer
	CreatedAt time.Time
}

func (AppUser) TableName() string {
	return "core.users"
}
