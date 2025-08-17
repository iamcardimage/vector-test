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

	// draft | submitted | approved | rejected | doc_requested | superseded_by_client_change
	Status string `gorm:"type:text;not null"`

	// Гибкие поля формы второй части
	Data datatypes.JSON `gorm:"type:jsonb"`

	// Риск второй части и срок следующей проверки
	RiskLevel string `gorm:"type:text"` // low | high
	DueAt     *time.Time

	// Акторы/аудит
	CreatedByUserID  *int
	UpdatedByUserID  *int
	ApprovedByUserID *int
	Reason           string `gorm:"type:text"`
}

func (SecondPartVersion) TableName() string {
	return "core.second_part_versions"
}
