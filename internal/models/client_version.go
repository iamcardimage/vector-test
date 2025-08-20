package models

import (
	"time"

	"gorm.io/datatypes"
)

type ClientVersion struct {
	ClientID     int `gorm:"not null;index"`
	Version      int `gorm:"not null"`
	Surname      string
	Name         string
	Patronymic   string
	Birthday     string
	BirthPlace   string
	ContactEmail string
	// ContactPhone   string
	// ContactAddress string
	// ContactCity    string
	// ContactState   string
	// ContactZip     string
	// ContactCountry string

	// риск из внешней базы (НЕ тот, что во второй части)
	ExternalRiskLevel string

	// хэш только по списку триггер-полей (для решения о 2-й части)
	SecondPartTriggerHash string `gorm:"not null"`

	NeedsSecondPart   bool           `gorm:"not null"`
	SecondPartCreated bool           `gorm:"not null"`
	Hash              string         `gorm:"not null"` // общий версионный хэш
	Status            string         // unchanged/changed
	Raw               datatypes.JSON `gorm:"type:jsonb"`
	SyncedAt          time.Time      `gorm:"not null"`
	ValidFrom         time.Time      `gorm:"not null"`
	ValidTo           *time.Time
	IsCurrent         bool `gorm:"not null;index"`
}

func (ClientVersion) TableName() string {
	return "core.clients_versions"
}
