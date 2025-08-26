// internal/models/contract.go
package models

import (
	"time"

	"gorm.io/datatypes"
)

type Contract struct {
	// Основные поля
	ID                         int `gorm:"primaryKey;column:id"`
	UserID                     int `gorm:"not null;index"`
	Comment                    *string
	CreatedAt                  time.Time `gorm:"not null"`
	UpdatedAt                  time.Time `gorm:"not null"`
	InnerCode                  string    `gorm:"not null;index"`
	IsPersonalInvestAccount    bool      `gorm:"not null"`
	IsPersonalInvestAccountNew bool      `gorm:"not null"`
	Kind                       string    `gorm:"not null;index"` // "broking", "depo"
	RialtoCode                 *string
	SignedAt                   *time.Time
	ClosedAt                   *time.Time
	Status                     string `gorm:"not null;index"` // "active", "closed"
	ContractOwnerType          string `gorm:"not null"`
	ContractOwnerID            int    `gorm:"not null"`
	Anketa                     *string
	OwnerID                    *int // parent contract
	CalculatedProfileID        *int
	DepoAccountsType           *string
	StrategyID                 *int
	StrategyName               *string
	TariffID                   *int
	TariffName                 *string
	UserLogin                  *string

	// JSON поля для полных данных
	Raw datatypes.JSON `gorm:"type:jsonb"`

	// Системные поля для синхронизации
	ExternalID int       `gorm:"not null;unique;index"`
	Hash       string    `gorm:"not null"` // для отслеживания изменений
	SyncedAt   time.Time `gorm:"not null"`
}

func (Contract) TableName() string {
	return "core.contracts"
}

// Добавляем к staging моделям
type StagingExternalContract struct {
	ID       int            `gorm:"primaryKey"`
	Raw      datatypes.JSON `gorm:"type:jsonb"`
	SyncedAt time.Time      `gorm:"not null"`
}

func (StagingExternalContract) TableName() string {
	return "staging.external_contracts"
}
