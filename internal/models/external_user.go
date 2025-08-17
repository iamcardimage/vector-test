package models

import "time"

type ExternalUser struct {
	ID         int `gorm:"primaryKey"`
	Surname    string
	Name       string
	Patronymic string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
