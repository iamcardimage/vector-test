package models

import (
	"time"

	"gorm.io/datatypes"
)

type SecondPartVersion struct {
	ClientID      int `gorm:"not null;index"`
	ClientVersion int `gorm:"not null;index"`

	Version   int       `gorm:"not null"`
	IsCurrent bool      `gorm:"not null;index"`
	ValidFrom time.Time `gorm:"not null"`
	ValidTo   *time.Time

	// draft | submitted | approved | rejected | doc_requested
	Status string `gorm:"type:text;not null"`

	Data datatypes.JSON `gorm:"type:jsonb"`

	RiskLevel string `gorm:"type:text"` // low | high
	DueAt     *time.Time

	CreatedByUserID  *int
	UpdatedByUserID  *int
	ApprovedByUserID *int
	Reason           string `gorm:"type:text"`
}

func (SecondPartVersion) TableName() string {
	return "core.second_part_versions"
}
