package models

import (
	"time"

	"gorm.io/datatypes"
)

type StagingExternalUser struct {
	ID       int            `gorm:"primaryKey"`
	Raw      datatypes.JSON `gorm:"type:jsonb;not null"`
	SyncedAt time.Time      `gorm:"not null"`
}

func (StagingExternalUser) TableName() string {
	return "staging.external_users"
}
