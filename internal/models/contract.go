package models

import (
	"time"

	"gorm.io/datatypes"
)

type Contract struct {
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
	OwnerID                    *int
	CalculatedProfileID        *int
	DepoAccountsType           *string
	StrategyID                 *int
	StrategyName               *string
	TariffID                   *int
	TariffName                 *string
	UserLogin                  *string

	Raw datatypes.JSON `gorm:"type:jsonb"`

	ExternalID int       `gorm:"not null;unique;index"`
	Hash       string    `gorm:"not null"`
	SyncedAt   time.Time `gorm:"not null"`
}

func (Contract) TableName() string {
	return "core.contracts"
}

type StagingExternalContract struct {
	ID       int            `gorm:"primaryKey"`
	Raw      datatypes.JSON `gorm:"type:jsonb"`
	SyncedAt time.Time      `gorm:"not null"`
}

func (StagingExternalContract) TableName() string {
	return "staging.external_contracts"
}
