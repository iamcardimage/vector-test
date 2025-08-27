package models

import (
	"time"

	"gorm.io/datatypes"
)

type SecondPartCheck struct {
	ID                uint           `gorm:"primaryKey"`
	ClientID          int            `gorm:"not null;index"`
	SecondPartVersion int            `gorm:"not null;index"`
	Kind              string         `gorm:"type:text;not null"`
	Status            string         `gorm:"type:text;not null"`
	Payload           datatypes.JSON `gorm:"type:jsonb"`
	Result            datatypes.JSON `gorm:"type:jsonb"`
	RunAt             time.Time      `gorm:"not null"`
	FinishedAt        *time.Time
	RunByUserID       *int
	CreatedAt         time.Time
}

func (SecondPartCheck) TableName() string {
	return "core.second_part_checks"
}
