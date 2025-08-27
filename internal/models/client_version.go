package models

import (
	"time"

	"gorm.io/datatypes"
)

type ClientVersion struct {
	ClientID       int `gorm:"not null;index"`
	Version        int `gorm:"not null"`
	Surname        string
	Name           string
	Patronymic     string
	Birthday       string
	BirthPlace     string
	ContactEmail   string
	Inn            string
	Snils          string
	CreatedLKAt    string
	UpdatedLKAt    string
	PassIssuerCode string
	PassSeries     string
	PassNumber     string
	PassIssueDate  string
	PassIssuer     string
	MainPhone      string

	ID                int    `gorm:"column:external_id;index"`
	Login             string `gorm:"index"`
	LockedAt          *time.Time
	CurrentSignInAt   *time.Time
	SignInCount       *int
	NeedToSetPassword *bool

	Blocked       *bool
	BlockedReason string
	BlockType     string

	Male               *bool
	IsRfResident       *bool
	DocumentType       string
	DocumentCountry    string
	LegalCapacity      string
	IsRfTaxpayer       *bool
	PifsPortfolioCode  *int
	ExternalIDStr      string `gorm:"column:external_id_str;index"` // external_id
	IsValidInfo        *bool
	QualifiedInvestor  *bool
	RiskLevel          string `gorm:"index"`
	FillStage          string
	IsFilled           *bool
	EsiaID             *int
	IdentificationType string
	AgentID            *int
	AgentPointID       *int
	TaxStatus          string
	IsAmericanNational *bool

	Country  string
	Region   string
	Index    *int
	City     string
	Street   string
	House    string
	Corps    *int
	Flat     *int
	District string

	SignatureType              string
	DataReceivedDigitalProfile *bool

	FromCompanySettings     datatypes.JSON `gorm:"type:jsonb"`
	Settings                datatypes.JSON `gorm:"type:jsonb"`
	PersonInfo              datatypes.JSON `gorm:"type:jsonb"`
	Manager                 datatypes.JSON `gorm:"type:jsonb"`
	Checks                  datatypes.JSON `gorm:"type:jsonb"`
	Note                    datatypes.JSON `gorm:"type:jsonb"`
	AdSource                datatypes.JSON `gorm:"type:jsonb"`
	SignatureAllowedNumbers datatypes.JSON `gorm:"type:jsonb"`

	ExternalRiskLevel string

	SecondPartTriggerHash string `gorm:"not null"`

	NeedsSecondPart   bool           `gorm:"not null"`
	SecondPartCreated bool           `gorm:"not null"`
	Hash              string         `gorm:"not null"`
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
