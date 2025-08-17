package models

import (
	"time"

	"gorm.io/datatypes"
)

type SecondPartCheck struct {
	ID                uint           `gorm:"primaryKey"`
	ClientID          int            `gorm:"not null;index"`
	SecondPartVersion int            `gorm:"not null;index"`     // к какой версии 2-й части относится
	Kind              string         `gorm:"type:text;not null"` // тип проверки (например, "passport", "risk", ...)
	Status            string         `gorm:"type:text;not null"` // pending|passed|failed
	Payload           datatypes.JSON `gorm:"type:jsonb"`         // входные параметры/контекст
	Result            datatypes.JSON `gorm:"type:jsonb"`         // результат внешней проверки
	RunAt             time.Time      `gorm:"not null"`
	FinishedAt        *time.Time
	RunByUserID       *int
	CreatedAt         time.Time
}

func (SecondPartCheck) TableName() string {
	return "core.second_part_checks"
}
