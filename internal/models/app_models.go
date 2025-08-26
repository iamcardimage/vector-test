package models

import (
	"time"
)

// ClientWithSP представляет клиента с информацией о второй части (для app сервиса)
type ClientWithSP struct {
	ClientID          int    `gorm:"column:client_id" json:"id"`
	Surname           string `json:"surname"`
	Name              string `json:"name"`
	Patronymic        string `json:"patronymic"`
	ExternalRiskLevel string `gorm:"column:external_risk_level" json:"external_risk_level"`
	NeedsSecondPart   bool   `gorm:"column:needs_second_part" json:"needs_second_part"`
	SecondPartCreated bool   `gorm:"column:second_part_created" json:"second_part_created"`

	ClientVersion   int        `gorm:"column:client_version" json:"client_version"`
	SpStatus        *string    `gorm:"column:sp_status" json:"sp_status,omitempty"`
	SpDueAt         *time.Time `gorm:"column:sp_due_at" json:"sp_due_at,omitempty"`
	SpClientVersion *int       `gorm:"column:sp_client_version" json:"sp_client_version,omitempty"`
}
